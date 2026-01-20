package render_runtime

import (
	"fmt"
	"image/color"
	"math"

	"github.com/tronbyt/pixlet/render"
	"go.starlark.net/starlark"
)

func DataPointElementFromStarlark(value starlark.Value) (float64, error) {
	if _, isNone := value.(starlark.NoneType); isNone {
		return math.NaN(), nil
	} else if result, isFloat := starlark.AsFloat(value); isFloat {
		return result, nil
	} else {
		return math.NaN(), fmt.Errorf("invalid type for data point element 0: %s (expected None or float)", value.Type())
	}
}

func DataPointFromStarlark(value starlark.Value) ([2]float64, error) {
	result := [2]float64{math.NaN(), math.NaN()}

	tuple, isTuple := value.(starlark.Tuple)
	if !isTuple {
		return result, fmt.Errorf("invalid type for data point: %s (expected a 2-tuple)", value.Type())
	} else if tuple.Len() == 0 {
		// (NaN, NaN)
		return result, nil
	} else if tuple.Len() != 2 {
		return result, fmt.Errorf("invalid type for data point: %s (expected a 2-tuple)", value.Type())
	}

	for i := range result {
		if value, err := DataPointElementFromStarlark(tuple.Index(i)); err == nil {
			result[i] = value
		} else {
			return result, err
		}
	}

	return result, nil
}

func DataSeriesFromStarlark(list *starlark.List) ([][2]float64, error) {
	result := make([][2]float64, 0)

	for i := range list.Len() {
		if val, err := DataPointFromStarlark(list.Index(i)); err == nil {
			result = append(result, val)
		} else {
			return nil, err
		}
	}

	return result, nil
}

func WeightsFromStarlark(list *starlark.List) ([]float64, error) {
	result := make([]float64, 0)

	for i := range list.Len() {
		if val, err := DataPointElementFromStarlark(list.Index(i)); err == nil {
			result = append(result, val)
		} else {
			return nil, err
		}
	}

	return result, nil
}

func ColorSeriesFromStarlark(list *starlark.List) ([]color.Color, error) {
	result := make([]color.Color, 0)

	for i := range list.Len() {
		c := list.Index(i)

		switch v := c.(type) {
		case starlark.String:
			c, err := render.ParseColor(v.GoString())
			if err != nil {
				return nil, fmt.Errorf("colors[%v] is not a valid hex string: %+v", i, c)
			}
			result = append(result, c)
		default:
			return nil, fmt.Errorf("colors[%v] is not a valid string", i)
		}
	}

	return result, nil
}

func VerticesFromStarlark(value starlark.Value) ([]render.Point, error) {
	iterable, ok := value.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("vertices must be iterable, found %s", value.Type())
	}

	n := starlark.Len(value)
	if n < 0 {
		n = 0
	}
	result := make([]render.Point, 0, n)

	iter := iterable.Iterate()
	defer iter.Done()
	var v starlark.Value
	for i := 0; iter.Next(&v); i++ {
		tuple, ok := v.(starlark.Tuple)
		if !ok || tuple.Len() != 2 {
			return nil, fmt.Errorf("vertex %d must be a 2-tuple (x, y)", i)
		}

		x, okX := starlark.AsFloat(tuple.Index(0))
		y, okY := starlark.AsFloat(tuple.Index(1))

		if !okX || !okY {
			return nil, fmt.Errorf("vertex %d coordinates must be numbers", i)
		}

		result = append(result, render.Point{X: x, Y: y})
	}

	return result, nil
}
