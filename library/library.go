//go:build lib

package main

/*
#include <stdbool.h>
#include <stdlib.h>

// When making breaking changes to the library API, increment libpixletAPIVersion in library/version.c.
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
	"unsafe"

	"github.com/tronbyt/pixlet/encode"
	"github.com/tronbyt/pixlet/runtime"
	"github.com/tronbyt/pixlet/runtime/modules/render_runtime/canvas"
	"github.com/tronbyt/pixlet/server/loader"
)

const (
	statusErrInvalidConfig  = -1
	statusErrRenderFailure  = -2
	statusErrInvalidFilters = -3
	statusErrHandlerFailure = -4
	statusErrInvalidPath    = -5
	statusErrStarSuffix     = -6
	statusErrUnknownApplet  = -7
)

// render_app renders an applet based on the provided parameters.
//
// Arguments:
//   - pathPtr (*C.char): A C string representing the path to the applet file or directory.
//   - configPtr (*C.char): A C string containing a JSON-encoded configuration map.
//   - width (C.int): The width of the rendered output.
//   - height (C.int): The height of the rendered output.
//   - maxDuration (C.int): The maximum duration (in milliseconds) for rendering.
//   - timeout (C.int): The timeout (in milliseconds) for rendering.
//   - imageFormat (C.int): The format of the rendered image (e.g., PNG, GIF).
//   - silenceOutput (C.int): A flag to suppress output (non-zero to silence).
//   - output2x (C.bool): Render at 2x resolution
//   - filtersPtr (*C.char): A JSON string for optional filters (e.g. {"magnify":2,"color_filter":"warm"})
//   - tzPtr (*C.char): The local timezone. Defaults to the system time.
//
// Returns:
//   - (*C.uchar): A pointer to the rendered image bytes.
//   - (C.int): The length of the rendered image bytes, or a negative status code on error.
//   - (*C.char): A pointer to a JSON-encoded array containing messages printed by the application.
//   - (*C.char): A pointer to an error message (if any).

//export render_app
func render_app(
	pathPtr *C.char,
	configPtr *C.char,
	width, height, maxDuration, timeout, imageFormat, silenceOutput C.int,
	output2x C.bool,
	filtersPtr, tzPtr *C.char,
) (*C.uchar, C.int, *C.char, *C.char) {
	path := C.GoString(pathPtr)
	configStr := C.GoString(configPtr)

	var config map[string]string
	if err := json.Unmarshal([]byte(configStr), &config); err != nil {
		return nil, C.int(statusErrInvalidConfig), nil, C.CString(fmt.Sprintf("error parsing config: %v", err))
	}

	var filters *encode.RenderFilters
	if filtersPtr != nil {
		var parsed encode.RenderFilters
		filtersStr := C.GoString(filtersPtr)
		if err := json.Unmarshal([]byte(filtersStr), &parsed); err != nil {
			return nil, C.int(statusErrInvalidFilters), nil, C.CString(fmt.Sprintf("invalid filters JSON: %v", err))
		}
		filters = &parsed
	}

	location := time.Local
	if tzPtr != nil {
		if tzRaw := C.GoString(tzPtr); tzRaw != "" {
			if v, err := time.LoadLocation(C.GoString(tzPtr)); err == nil {
				location = v
			}
		}
	}

	meta := canvas.Metadata{
		Width:  int(width),
		Height: int(height),
		Is2x:   bool(output2x),
	}

	result, messages, err := loader.RenderApplet(path, config, meta, int(maxDuration), int(timeout), loader.ImageFormat(imageFormat), silenceOutput != 0, location, filters)

	messagesJSON, _ := json.Marshal(messages)
	if err != nil {
		return nil, C.int(statusErrRenderFailure), C.CString(string(messagesJSON)), C.CString(fmt.Sprintf("error rendering: %v", err))
	}

	return (*C.uchar)(C.CBytes(result)), C.int(len(result)), C.CString(string(messagesJSON)), nil
}

func errorStatus(err error) int {
	if err != nil {
		var pathErr *os.PathError
		switch {
		case errors.As(err, &pathErr):
			return statusErrInvalidPath
		case errors.Is(err, runtime.ErrStarSuffix):
			return statusErrStarSuffix
		}
	}
	return statusErrUnknownApplet
}

//export get_schema
func get_schema(pathPtr *C.char, width, height C.int, output2x C.bool) (*C.char, C.int) {
	path := C.GoString(pathPtr)

	meta := canvas.Metadata{
		Width:  int(width),
		Height: int(height),
		Is2x:   bool(output2x),
	}

	applet, err := runtime.NewAppletFromPath(path, runtime.WithCanvasMeta(meta))
	if err != nil {
		status := errorStatus(err)
		return nil, C.int(status)
	}
	defer applet.Close()

	return (*C.char)(C.CString(string(applet.SchemaJSON))), 0
}

//export call_handler
func call_handler(
	pathPtr, configPtr *C.char,
	width, height C.int,
	output2x C.bool,
	handlerName, parameter *C.char,
) (*C.char, C.int, *C.char) {
	path := C.GoString(pathPtr)
	configStr := C.GoString(configPtr)

	var config map[string]string
	if configStr != "" {
		if err := json.Unmarshal([]byte(configStr), &config); err != nil {
			return nil, C.int(statusErrInvalidConfig), C.CString(fmt.Sprintf("error parsing config: %v", err))
		}
	}

	meta := canvas.Metadata{
		Width:  int(width),
		Height: int(height),
		Is2x:   bool(output2x),
	}

	applet, err := runtime.NewAppletFromPath(path, runtime.WithCanvasMeta(meta))
	if err != nil {
		status := errorStatus(err)
		return nil, C.int(status), C.CString(fmt.Sprintf("error loading app: %v", err))
	}
	defer applet.Close()

	result, err := applet.CallSchemaHandler(context.Background(), C.GoString(handlerName), C.GoString(parameter), config)
	if err != nil {
		return nil, C.int(statusErrHandlerFailure), C.CString(err.Error())
	}

	return (*C.char)(C.CString(result)), 0, nil
}

//export free_bytes
func free_bytes(ptr *C.uchar) {
	C.free(unsafe.Pointer(ptr))
}

//export init_cache
func init_cache() {
	cache := runtime.NewInMemoryCache()
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)
}

//export init_redis_cache
func init_redis_cache(redisURL *C.char) {
	cache := runtime.NewRedisCache(C.GoString(redisURL))
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)
}

func main() {}
