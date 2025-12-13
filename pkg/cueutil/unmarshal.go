package cueutil

import (
	"os"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/cuecontext"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
)

type UnmarshalOptions struct {
	Map      map[string]any
	Validate bool
	// Options
	BuildOptions    []cue.BuildOption
	EncodeOptions   []cue.EncodeOption
	ValidateOptions []cue.Option
	SyntaxOptions   []cue.Option
}

type Unmarshal[T any] struct {
	UnmarshalOptions
}

// NewDefaultUnmarshal returns a default Unmarshal struct.
func NewDefaultUnmarshal[T any]() *Unmarshal[T] {
	return &Unmarshal[T]{
		UnmarshalOptions: UnmarshalOptions{
			Map:      nil,
			Validate: true,
		},
	}
}

// NewUnmarshal returns an Unmarshal[T] with the provided opts.
func NewUnmarshal[T any](opts UnmarshalOptions) *Unmarshal[T] {
	return &Unmarshal[T]{opts}
}

// Compile sets the receivers compile to the provided map.
func (u *Unmarshal[T]) Compile(c map[string]any) {
	u.Map = c
}

// NewFile unmarshals a file from the specified path, returning the data.
func (u *Unmarshal[T]) NewFile(path string) (T, *ast.Node, error) {
	var t T

	data, err := os.ReadFile(path)
	if err != nil {
		return t, nil, errutil.New("os.ReadFile", err)
	}

	f, err := u.Bytes(data, &t)

	return t, f, err
}

// NewBytes unmarshals the specified byte slice, returning the data.
func (u *Unmarshal[T]) NewBytes(data []byte) (T, *ast.Node, error) {
	var t T

	f, err := u.Bytes(data, &t)

	return t, f, err
}

// File unmarshals a file from the specified path into T.
func (u *Unmarshal[T]) File(path string, t *T) (*ast.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errutil.New("os.ReadFile", err)
	}

	f, err := u.Bytes(data, t)

	return f, err
}

// Bytes unmarshals the specified byte slice into T.
//
// ast.Node is passed instead of an ast.File or ast.StructLit
// because it can change depending on what is passed to
// ctx.Compile. Due to this, we only check if the type
// matches any of the types we know it can be (ast.File, ast.StructLit).
func (u *Unmarshal[T]) Bytes(data []byte, t *T) (*ast.Node, error) {
	ctx := cuecontext.New()

	v := ctx.CompileBytes(data, u.BuildOptions...)
	if v.Err() != nil {
		return nil, errutil.New("ctx.CompileBytes", v.Err())
	}

	if u.Map != nil {
		v = v.Unify(ctx.Encode(u.Map, u.EncodeOptions...))
		if v.Err() != nil {
			return nil, errutil.New("v.Unify", v.Err())
		}
	}

	if u.Validate {
		err := v.Validate(u.ValidateOptions...)
		if err != nil {
			return nil, errutil.New("v.Validate", err)
		}
	}

	if err := v.Decode(t); err != nil {
		return nil, errutil.New("v.Decode", err)
	}

	syn := v.Syntax(u.SyntaxOptions...)
	if err := ensureExpectedType(syn); err != nil {
		return nil, errutil.New("v.Syntax", err)
	}

	return &syn, nil
}

// ensureExpectedType checks if n matches the expected types, returning an error if not.
func ensureExpectedType(n ast.Node) error {
	switch n.(type) {
	case *ast.File, *ast.StructLit:
		logutil.Debugf(os.Stderr, "syntax type is %T\n", n)
		return nil
	default:
		return errutil.WithFramef("unexpected syntax type: %T", n)
	}
}
