package pipeline

import (
	"os"
	"slices"
	"sync"

	"github.com/ricochhet/pkg/logutil"
)

type Step[T any] struct {
	Name    string
	Aliases []string
	Step    func(*logutil.Logger, *T) (*T, error)
}

type Pipeline[T any] struct {
	mu    sync.Mutex
	steps []*Step[T]
}

// NewPipeline returns an empty Pipeline[T].
func NewPipeline[T any]() *Pipeline[T] {
	return &Pipeline[T]{}
}

// All returns the steps.
func (p *Pipeline[T]) All() []*Step[T] {
	p.mu.Lock()
	defer p.mu.Unlock()

	return slices.Clone(p.steps) // Return a copy.
}

// SetAll sets the steps.
func (p *Pipeline[T]) SetAll(list []*Step[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.steps = slices.Clone(list)
}

// Add appends to the steps.
func (p *Pipeline[T]) Add(funcs ...*Step[T]) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.steps = append(p.steps, funcs...)
}

// CopyFrom sets the steps to the target.
func (p *Pipeline[T]) CopyFrom(target *Pipeline[T]) {
	p.SetAll(target.All())
}

// Start runs the steps specified by the names, return T (final) based off T (initial).
func (p *Pipeline[T]) Start(initial *T, skip func([]string) bool) *T {
	p.mu.Lock()
	defer p.mu.Unlock()

	cur := initial

	for _, item := range p.steps {
		if skip(append(item.Aliases, item.Name)) {
			continue
		}

		// automate logger creation.
		logutil.MaxProcNameLength.Store(int32(len(item.Name)))

		t, err := item.Step(
			logutil.CreateLogger(item.Name, len(item.Name)%len(logutil.Colors)),
			cur)
		if err != nil {
			logutil.Errorf(os.Stderr, "Step returned an error: %v\n", err)
			continue
		}

		if t != nil {
			cur = t
		}
	}

	return cur
}
