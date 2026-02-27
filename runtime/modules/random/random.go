package random

import (
	cryptorand "crypto/rand"
	"fmt"
	"math/big"
	"math/rand/v2"
	"sync"

	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName    = "random"
	threadRandKey = "github.com/tronbyt/pixlet/runtime/random"
)

var (
	once   sync.Once
	module starlark.StringDict
)

func AttachToThread(t *starlark.Thread) {
	source := rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	t.SetLocal(threadRandKey, source)
}

func LoadModule() (starlark.StringDict, error) {
	once.Do(func() {
		module = starlark.StringDict{
			ModuleName: &starlarkstruct.Module{
				Name: ModuleName,
				Members: starlark.StringDict{
					"number": starlark.NewBuiltin("number", randomNumber),
					"seed":   starlark.NewBuiltin("seed", randomSeed),
				},
			},
		}
	})

	return module, nil
}

func randomSeed(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var starSeed starlark.Int

	if err := starlark.UnpackArgs(
		"seed",
		args, kwargs,
		"seed", &starSeed,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for seed: %w", err)
	}

	seed, err := starlarkutil.AsInt64(starSeed)
	if err != nil {
		return nil, fmt.Errorf("parsing seed: %w", err)
	}

	source := rand.New(rand.NewPCG(uint64(seed), uint64(seed)))
	thread.SetLocal(threadRandKey, source)

	return starlark.None, nil
}

func randomNumber(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		starMin starlark.Int
		starMax starlark.Int
		secure  starlark.Bool
	)

	if err := starlark.UnpackArgs(
		"number",
		args, kwargs,
		"min", &starMin,
		"max", &starMax,
		"secure?", &secure,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for random number: %w", err)
	}

	minVal, err := starlarkutil.AsInt64(starMin)
	if err != nil {
		return nil, fmt.Errorf("parsing min: %w", err)
	}

	maxVal, err := starlarkutil.AsInt64(starMax)
	if err != nil {
		return nil, fmt.Errorf("parsing max: %w", err)
	}

	if minVal < 0 {
		return nil, fmt.Errorf("min has to be 0 or greater")
	}

	if maxVal < minVal {
		return nil, fmt.Errorf("max is less than min")
	}

	shiftedMax := maxVal - minVal + 1

	var r int64
	if secure {
		v, err := cryptorand.Int(cryptorand.Reader, big.NewInt(shiftedMax))
		if err != nil {
			return nil, fmt.Errorf("reading random number: %w", err)
		}

		r = v.Int64()
	} else {
		rng, ok := thread.Local(threadRandKey).(*rand.Rand)
		if !ok || rng == nil {
			return nil, fmt.Errorf("RNG not set (very bad!)")
		}

		r = rng.Int64N(shiftedMax)
	}

	return starlark.MakeInt64(r + minVal), nil
}
