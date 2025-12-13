package config

import "sync"

type Exports struct {
	Achievements []byte
	Codex        []byte
	Customs      []byte
	Enemies      []byte
	Resources    []byte
	Virtuals     []byte
	Flavor       []byte
	Regions      []byte
	Weapons      []byte
	Warframes    []byte
	Sentinels    []byte
	AllScans     []byte
}

type ExportManager struct {
	mu      sync.Mutex
	exports *Exports
}

// NewExportManager creates an empty ExportManager.
func NewExportManager() *ExportManager {
	return &ExportManager{}
}

// All returns the exports.
func (m *ExportManager) All() *Exports {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.exports
}

// SetAll sets the exports.
func (m *ExportManager) SetAll(exports *Exports) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.exports = exports
}

// CopyFrom sets all exports to the target.
func (m *ExportManager) CopyFrom(target *ExportManager) {
	m.SetAll(target.All())
}
