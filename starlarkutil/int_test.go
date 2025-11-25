package starlarkutil

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestAsInt64(t *testing.T) {
	type args struct {
		x starlark.Int
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr require.ErrorAssertionFunc
	}{
		{"zero", args{starlark.MakeInt64(0)}, 0, require.NoError},
		{"positive", args{starlark.MakeInt64(1)}, 1, require.NoError},
		{"negative", args{starlark.MakeInt64(-1)}, -1, require.NoError},
		{"overflow", args{starlark.MakeInt64(math.MaxInt64).Add(starlark.MakeInt64(1))}, 0, require.Error},
		{"underflow", args{starlark.MakeInt64(math.MinInt64).Sub(starlark.MakeInt64(1))}, 0, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsInt64(tt.args.x)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAsUint64(t *testing.T) {
	type args struct {
		x starlark.Int
	}
	tests := []struct {
		name    string
		args    args
		want    uint64
		wantErr require.ErrorAssertionFunc
	}{
		{"zero", args{starlark.MakeUint64(0)}, 0, require.NoError},
		{"positive", args{starlark.MakeUint64(1)}, 1, require.NoError},
		{"negative", args{starlark.MakeInt64(-1)}, 0, require.Error},
		{"overflow", args{starlark.MakeUint64(math.MaxUint64).Add(starlark.MakeUint(1))}, 0, require.Error},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AsUint64(tt.args.x)
			tt.wantErr(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
