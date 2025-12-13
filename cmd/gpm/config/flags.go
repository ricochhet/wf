package config

import (
	"fmt"
	"os"
	"strconv"
)

type Flags struct {
	Taskfile       string `json:"taskfile"`
	Dotfile        string `json:"dotfile"`
	Envfile        string `json:"envfile"`
	EnvOverload    bool   `json:"envOverload"`
	Port           uint   `json:"port"`
	StartRPCServer bool   `json:"startRpcServer"`
	BaseDir        string `json:"basedir"`
	BasePort       uint   `json:"baseport"`
	SetPorts       bool   `json:"setPorts"`
	RestartOnError bool   `json:"restartOnError"`
	ExitOnError    bool   `json:"exitOnError"`
	ExitOnStop     bool   `json:"exitOnStop"`
	LogTime        bool   `json:"logTime"`
	Pty            bool   `json:"pty"`
	Interval       uint   `json:"interval"`
	ReverseOnStop  bool   `json:"reverseOnStop"`
	InheritStdin   bool   `json:"inheritStdin"`
	// Internals.
	Args      []string `json:"args"`
	Envfiles  []string `json:"envfiles"`
	VarPasses int      `json:"vPasses"`
	Global    string   `json:"global"`
	Debug     bool     `json:"debug"`
	QuickEdit bool     `json:"quickEdit"` // noop on non-Windows systems.
	Optionals bool     `json:"optionals"`
}

// DefaultServer returns the default RPC address:port.
func DefaultServer(serverPort uint) string {
	if s, ok := os.LookupEnv("GPM_RPC_SERVER"); ok {
		return s
	}

	if serverPort == 0 {
		serverPort = DefaultPort()
	}

	return fmt.Sprintf("127.0.0.1:%d", serverPort)
}

// DefaultAddr returns the default RPC address.
func DefaultAddr() string {
	if s, ok := os.LookupEnv("GPM_RPC_ADDR"); ok {
		return s
	}

	return "0.0.0.0"
}

// DefaultPort returns the default RPC port.
func DefaultPort() uint {
	s := os.Getenv("GPM_RPC_PORT")
	if s != "" {
		i, err := strconv.Atoi(s)
		if err == nil {
			return uint(i)
		}
	}

	return 8555
}
