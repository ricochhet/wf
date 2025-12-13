package config

import (
	"os/exec"
	"slices"
	"sync"
)

type ProcInfo struct {
	Name       string
	Desc       string
	Aliases    []string
	Cmdline    []string
	Cmd        *exec.Cmd
	Env        map[string][]string
	Steps      []string
	Dir        string
	Fork       bool
	Port       uint
	Silent     bool
	SetPort    bool
	ColorIndex int

	// True if we called stopProc to kill the process, in which case an
	// *os.ExitError is not the fault of the subprocess
	StoppedBySupervisor bool
	RestartOnError      bool
	InheritStdin        bool

	Mu      sync.Mutex
	Cond    *sync.Cond
	WaitErr error
}

type ProcManager struct {
	mu   sync.Mutex
	list []*ProcInfo
}

// NewProcManager creates an empty ProcManager.
func NewProcManager() *ProcManager {
	return &ProcManager{}
}

// All returns the list.
func (m *ProcManager) All() []*ProcInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	return slices.Clone(m.list) // Return a copy.
}

// SetAll sets the list.
func (m *ProcManager) SetAll(list []*ProcInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.list = slices.Clone(list)
}

// Add appends to the list.
func (m *ProcManager) Add(proc ...*ProcInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.list = append(m.list, proc...)
}

// CopyFrom sets the list to the target.
func (m *ProcManager) CopyFrom(target *ProcManager) {
	m.SetAll(target.All())
}
