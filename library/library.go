//go:build lib

package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unsafe"

	"tidbyt.dev/pixlet/runtime"
	"tidbyt.dev/pixlet/server/loader"
	"tidbyt.dev/pixlet/tools"
)

// render_app renders an applet based on the provided parameters.
//
// Arguments:
//   - pathPtr (*C.char): A C string representing the path to the applet file or directory.
//   - configPtr (*C.char): A C string containing a JSON-encoded configuration map.
//   - width (C.int): The width of the rendered output.
//   - height (C.int): The height of the rendered output.
//   - magnify (C.int): The magnification level for the rendered output.
//   - maxDuration (C.int): The maximum duration (in milliseconds) for rendering.
//   - timeout (C.int): The timeout (in milliseconds) for rendering.
//   - imageFormat (C.int): The format of the rendered image (e.g., PNG, GIF).
//   - silenceOutput (C.int): A flag to suppress output (non-zero to silence).
//
// Returns:
//   - (*C.uchar): A pointer to the rendered image bytes.
//   - (C.int): The length of the rendered image bytes, or a negative status code on error.
//   - (*C.char): A pointer to a JSON-encoded array containing messages printed by the application.
//   - (*C.char): A pointer to an error message (if any).

//export render_app
func render_app(pathPtr *C.char, configPtr *C.char, width, height, magnify, maxDuration, timeout, imageFormat, silenceOutput C.int) (*C.uchar, C.int, *C.char, *C.char) {
	path := C.GoString(pathPtr)
	configStr := C.GoString(configPtr)

	var config map[string]string
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		return nil, -1, nil, C.CString(fmt.Sprintf("error parsing config: %v", err))
	}

	result, messages, err := loader.RenderApplet(path, config, int(width), int(height), int(magnify), int(maxDuration), int(timeout), loader.ImageFormat(imageFormat), silenceOutput != 0)
	messagesJSON, _ := json.Marshal(messages)
	if err != nil {
		return nil, -2, C.CString(string(messagesJSON)), C.CString(fmt.Sprintf("error rendering: %v", err))
	}
	return (*C.uchar)(C.CBytes(result)), C.int(len(result)), C.CString(string(messagesJSON)), nil
}

func appletFromPath(path string) (*runtime.Applet, int) {
	// check if path exists, and whether it is a directory or a file
	info, err := os.Stat(path)
	if err != nil {
		return nil, -1
	}

	var fs fs.FS
	if info.IsDir() {
		fs = os.DirFS(path)
	} else {
		if !strings.HasSuffix(path, ".star") {
			return nil, -2
		}

		fs = tools.NewSingleFileFS(path)
	}

	applet, err := runtime.NewAppletFromFS(filepath.Base(path), fs)
	if err != nil {
		return nil, -3
	}

	return applet, 0
}

//export get_schema
func get_schema(pathPtr *C.char) (*C.char, C.int) {
	path := C.GoString(pathPtr)

	applet, status := appletFromPath(path)
	if status != 0 {
		return nil, C.int(status)
	}

	return (*C.char)(C.CString(string(applet.SchemaJSON))), 0
}

//export call_handler
func call_handler(pathPtr, handlerName, parameter *C.char) (*C.char, C.int) {
	path := C.GoString(pathPtr)

	applet, status := appletFromPath(path)
	if status != 0 {
		return nil, C.int(status)
	}

	result, err := applet.CallSchemaHandler(context.Background(), C.GoString(handlerName), C.GoString(parameter))
	if err != nil {
		return nil, -1
	}

	return (*C.char)(C.CString(result)), 0
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
