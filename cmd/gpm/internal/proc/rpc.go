package proc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"sync"
	"time"

	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/errutil"
)

// Gpm RPC server.
type Gpm struct {
	rpcChan chan<- *RPCMessage
	ctx     *Context
}

type RPCMessage struct {
	Msg  string
	Args []string
	// Sending error (if any) when the task completes.
	ErrCh chan error
}

// Start do start.
func (r *Gpm) Start(args []string, _ *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	for _, arg := range args {
		if err = r.ctx.StartProc(arg, nil, nil); err != nil {
			break
		}
	}

	return errutil.WithFrame(err)
}

// Stop do stop.
func (r *Gpm) Stop(args []string, _ *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	errChan := make(chan error, 1)
	r.rpcChan <- &RPCMessage{
		Msg:   "stop",
		Args:  args,
		ErrCh: errChan,
	}

	err = <-errChan

	return err
}

// StopAll do stop all.
func (r *Gpm) StopAll(_ []string, _ *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	for _, proc := range r.ctx.SharedProc.All() {
		if err = r.ctx.StopProc(proc.Name, nil); err != nil {
			break
		}
	}

	return errutil.WithFrame(err)
}

// Restart do restart.
func (r *Gpm) Restart(args []string, _ *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	for _, arg := range args {
		if err = r.ctx.RestartProc(arg); err != nil {
			break
		}
	}

	return errutil.WithFrame(err)
}

// RestartAll do restart all.
func (r *Gpm) RestartAll(_ []string, _ *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	for _, proc := range r.ctx.SharedProc.All() {
		if err = r.ctx.RestartProc(proc.Name); err != nil {
			break
		}
	}

	return errutil.WithFrame(err)
}

// List do list.
func (r *Gpm) List(_ []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	*ret = ""
	for _, proc := range r.ctx.SharedProc.All() {
		*ret += proc.Name + "\n"
	}

	return errutil.WithFrame(err)
}

// Status do status.
func (r *Gpm) Status(_ []string, ret *string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()

	*ret = ""

	for _, proc := range r.ctx.SharedProc.All() {
		if proc.Cmd != nil {
			*ret += "*" + proc.Name + "\n"
		} else {
			*ret += " " + proc.Name + "\n"
		}
	}

	return errutil.WithFrame(err)
}

// command: run.
func Run(cmd string, args []string, serverPort uint) error {
	client, err := rpc.Dial("tcp", config.DefaultServer(serverPort))
	if err != nil {
		return errutil.New("rpc.Dial", err)
	}

	defer client.Close()

	var ret string

	switch cmd {
	case "start":
		return client.Call("Gpm.Start", args, &ret)
	case "stop":
		return client.Call("Gpm.Stop", args, &ret)
	case "stop-all":
		return client.Call("Gpm.StopAll", args, &ret)
	case "restart":
		return client.Call("Gpm.Restart", args, &ret)
	case "restart-all":
		return client.Call("Gpm.RestartAll", args, &ret)
	case "list":
		if err := client.Call("Gpm.List", args, &ret); err != nil {
			return errutil.New("client.Call (Gpm.List)", err)
		}

		fmt.Print(ret)

		return nil
	case "status":
		if err := client.Call("Gpm.Status", args, &ret); err != nil {
			return errutil.New("client.Call (Gpm.Status)", err)
		}

		fmt.Print(ret)

		return nil
	}

	return errors.New("unknown command")
}

// StartServer starts the RPC server.
func (ctx *Context) StartServer(
	rpcCtx context.Context,
	rpcChan chan<- *RPCMessage,
	listenPort uint,
) error {
	gm := &Gpm{
		rpcChan: rpcChan,
		ctx:     ctx,
	}
	if err := rpc.Register(gm); err != nil {
		return errutil.New("rpc.Register", err)
	}

	lc := net.ListenConfig{}

	server, err := lc.Listen(rpcCtx, "tcp", fmt.Sprintf("%s:%d", config.DefaultAddr(), listenPort))
	if err != nil {
		return errutil.New("net.Listen", err)
	}

	var wg sync.WaitGroup

	acceptingConns := true

outer:
	for acceptingConns {
		conns := make(chan net.Conn, 1)

		go func() {
			conn, err := server.Accept()
			if err != nil {
				return
			}

			conns <- conn
		}()

		select {
		case <-rpcCtx.Done():
			acceptingConns = false //nolint:ineffassign,wastedassign // wontfix
			break outer
		case client := <-conns: // Server is not canceled.
			wg.Add(1)

			go func() {
				defer wg.Done()

				rpc.ServeConn(client)
			}()
		}
	}

	done := make(chan struct{}, 1)

	go func() {
		wg.Wait()

		done <- struct{}{}
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("RPC server did not shut down in 10 seconds, quitting")
	}
}
