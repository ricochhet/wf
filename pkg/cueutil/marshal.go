package cueutil

import (
	"bytes"
	"fmt"
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/format"
	"github.com/ricochhet/pkg/errutil"
)

type MarshalOptions struct {
	Package  string
	Comp     map[string]any
	Validate bool
	// Options
	EncodeOptions     []cue.EncodeOption
	CompEncodeOptions []cue.EncodeOption
	ValidateOptions   []cue.Option
	FormatOptions     []format.Option
	SyntaxOptions     []cue.Option
}

type Marshal[T any] struct {
	MarshalOptions
}

// NewDefaultMarshal returns a default Marshal struct.
func NewDefaultMarshal[T any]() *Marshal[T] {
	return &Marshal[T]{
		MarshalOptions: MarshalOptions{
			Package:  "",
			Comp:     nil,
			Validate: true,
		},
	}
}

// NewMarshal returns a Marshal[T] with the provided opts.
func NewMarshal[T any](opts MarshalOptions) *Marshal[T] {
	return &Marshal[T]{opts}
}

// File marshals t into a file.
func (m *Marshal[T]) File(path string, t T) error {
	data, err := m.Bytes(t)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// Bytes marshals t into a new byte slice.
func (m *Marshal[T]) Bytes(t T) ([]byte, error) {
	ctx := cuecontext.New()

	v := ctx.Encode(t, m.EncodeOptions...)
	if v.Err() != nil {
		return nil, errutil.New("ctx.Encode", v.Err())
	}

	if m.Comp != nil {
		v = v.Unify(ctx.Encode(m.Comp, m.CompEncodeOptions...))
		if v.Err() != nil {
			return nil, errutil.New("v.Unify", v.Err())
		}
	}

	if m.Validate {
		if err := v.Validate(m.ValidateOptions...); err != nil {
			return nil, errutil.New("v.Validate", err)
		}
	}

	syn := v.Syntax(m.ensureSyntaxOpts()...)

	body, err := format.Node(syn, m.FormatOptions...)
	if err != nil {
		return nil, errutil.New("format.Node", err)
	}

	var b bytes.Buffer

	if m.Package != "" {
		fmt.Fprintf(&b, "package %s\n\n", m.Package)
	}

	b.Write(body)

	return body, nil
}

// ensureSyntaxOpts returns a default set of syntax options if none are specified in the receiver.
func (m *Marshal[T]) ensureSyntaxOpts() []cue.Option {
	if m.SyntaxOptions == nil {
		return []cue.Option{
			cue.Final(),
			cue.Concrete(true),
			cue.Definitions(true),
			cue.Docs(true),
		}
	}

	return m.SyntaxOptions
}
