package file

import (
	"errors"
	"fmt"
	"io/fs"
	"unsafe"

	"github.com/mitchellh/hashstructure/v2"
	"go.starlark.net/starlark"
)

type File struct {
	FS   fs.FS
	Path string
}

func (f *File) readall(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var modeParam starlark.String

	if err := starlark.UnpackArgs(
		b.Name(), args, kwargs,
		"mode?", &modeParam,
	); err != nil {
		return nil, err
	}

	mode, err := ParseMode(string(modeParam))
	if err != nil {
		return nil, err
	}

	bs, err := fs.ReadFile(f.FS, f.Path)
	if err != nil {
		return nil, err
	}

	switch mode {
	case ModeReadText:
		str := unsafe.String(unsafe.SliceData(bs), len(bs))
		return starlark.String(str), nil
	case ModeReadBinary:
		return starlark.Bytes(bs), nil
	default:
		return nil, ErrUnsupportedMode
	}
}

func (f *File) AttrNames() []string {
	return []string{"path", "readall"}
}

func (f *File) Attr(name string) (starlark.Value, error) {
	switch name {
	case "path":
		return starlark.String(f.Path), nil
	case "readall":
		return starlark.NewBuiltin("readall", f.readall), nil
	default:
		return nil, nil
	}
}

func (f *File) String() string       { return "File(...)" }
func (f *File) Type() string         { return "File" }
func (f *File) Freeze()              {}
func (f *File) Truth() starlark.Bool { return true }

func (f *File) Hash() (uint32, error) {
	sum, err := hashstructure.Hash(f, hashstructure.FormatV2, nil)
	return uint32(sum), err
}

type Mode uint8

const (
	ModeNone Mode = iota
	ModeReadText
	ModeReadBinary
)

var ErrUnsupportedMode = errors.New("unsupported mode")

func ParseMode(mode string) (Mode, error) {
	switch mode {
	case "", "r", "rt":
		return ModeReadText, nil
	case "rb":
		return ModeReadBinary, nil
	default:
		return ModeNone, fmt.Errorf("%w: %s", ErrUnsupportedMode, mode)
	}
}
