package proc

import (
	"errors"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
	"golang.org/x/sys/windows"
)

var (
	cmdStart  = []string{"cmd", "/c"}
	procAttrs = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_UNICODE_ENVIRONMENT | windows.CREATE_NEW_PROCESS_GROUP,
	}
	forkProcAttrs = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NEW_PROCESS_GROUP | windows.DETACHED_PROCESS,
	}
)

// terminateProc terminates the process by sending the signal to the process.
func terminateProc(proc *config.ProcInfo, _ os.Signal) error {
	dll, err := windows.LoadDLL("kernel32.dll")
	if err != nil {
		return errutil.New("windows.LoadDLL (kernel32.dll)", err)
	}

	defer func() {
		if err := dll.Release(); err != nil {
			logutil.Errorf(os.Stderr, "Failed to release DLL: %v\n", err)
		}
	}()

	pid := proc.Cmd.Process.Pid

	f, err := dll.FindProc("AttachConsole")
	if err != nil {
		return errutil.New("dll.FindProc (AttachConsole)", err)
	}

	r1, _, err := f.Call(uintptr(pid))
	if r1 == 0 && !errors.Is(err, syscall.ERROR_ACCESS_DENIED) {
		return errutil.New("f.Call (pid)", err)
	}

	f, err = dll.FindProc("SetConsoleCtrlHandler")
	if err != nil {
		return errutil.New("dll.FindProc (SetConsoleCtrlHandler)", err)
	}

	r1, _, err = f.Call(0, 1)
	if r1 == 0 {
		return errutil.New("f.Call (0, 1)", err)
	}

	f, err = dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return errutil.New("dll.FindProc (GenerateConsoleCtrlEvent)", err)
	}

	r1, _, err = f.Call(windows.CTRL_BREAK_EVENT, uintptr(pid))
	if r1 == 0 {
		return errutil.New("f.Call (CTRL_BREAK_EVENT)", err)
	}

	r1, _, err = f.Call(windows.CTRL_C_EVENT, uintptr(pid))
	if r1 == 0 {
		return errutil.New("f.Call (CTRL_C_EVENT)", err)
	}

	return nil
}

// killProc kills the proc with pid, as well as its children.
func killProc(process *os.Process) error {
	return process.Kill()
}

// NotifyCh create the terminate/interrupt notifier.
func NotifyCh() (<-chan os.Signal, func()) {
	sc := make(chan os.Signal, 10)
	signal.Notify(sc, os.Interrupt)

	return sc, func() {
		signal.Stop(sc)
	}
}

// startPTY starts a PTY (posix only).
func (ctx *Context) startPTY(logger *logutil.Logger, cmd *exec.Cmd) (func(), error) {
	cmd.Stdout = logger
	cmd.Stderr = logger

	return func() {}, nil
}
