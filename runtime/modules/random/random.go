package random

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/tronbyt/pixlet/starlarkutil"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName       = "random"
	threadRandKey    = "github.com/tronbyt/pixlet/runtime/random"
	randomSeedWindow = 15
)

var (
	once   sync.Once
	module starlark.StringDict
)

func AttachToThread(t *starlark.Thread) {
	nowSeconds := time.Now().UnixMilli() / 1000

	t.SetLocal(
		threadRandKey,
		rand.New(
			// Seed RNG with a constant for brief time
			// windows. This allows app to be "random",
			// while still enabling Tidbyt's backend to
			// cache the results.
			rand.NewSource(nowSeconds/randomSeedWindow),
		),
	)
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

	rng, ok := thread.Local(threadRandKey).(*rand.Rand)
	if !ok || rng == nil {
		return nil, fmt.Errorf("RNG not set (very bad)")
	}

	rng.Seed(seed)

	return starlark.None, nil
}

func randomNumber(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		starMin starlark.Int
		starMax starlark.Int
	)

	if err := starlark.UnpackArgs(
		"number",
		args, kwargs,
		"min", &starMin,
		"max", &starMax,
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

	rng, ok := thread.Local(threadRandKey).(*rand.Rand)
	if !ok || rng == nil {
		return nil, fmt.Errorf("RNG not set (very bad!)")
	}

	return starlark.MakeInt64(rng.Int63n(maxVal-minVal+1) + minVal), nil
}
