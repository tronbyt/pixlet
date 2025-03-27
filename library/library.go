//go:build lib

package main

/*
#include <stdlib.h>
*/
import "C"

import (
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

//export render_app
func render_app(pathPtr *C.char, configPtr *C.char, width, height, magnify, maxDuration, timeout, imageFormat, silenceOutput C.int) (*C.uchar, C.int) {
	path := C.GoString(pathPtr)
	configStr := C.GoString(configPtr)

	var config map[string]string
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		fmt.Printf("error parsing config: %v\n", err)
		return nil, -1
	}

	result, err := loader.RenderApplet(path, config, int(width), int(height), int(magnify), int(maxDuration), int(timeout), loader.ImageFormat(imageFormat), silenceOutput != 0)
	if err != nil {
		fmt.Printf("error rendering: %v\n", err)
		return nil, -2
	}
	return (*C.uchar)(C.CBytes(result)), C.int(len(result))
}

//export get_schema
func get_schema(pathPtr *C.char) (*C.uchar, C.int) {
	path := C.GoString(pathPtr)

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

	return (*C.uchar)(C.CBytes(applet.SchemaJSON)), C.int(len(applet.SchemaJSON))
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
