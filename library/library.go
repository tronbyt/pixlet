//go:build lib

package main

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
	"fmt"
	"unsafe"

	"tidbyt.dev/pixlet/runtime"
	"tidbyt.dev/pixlet/server/loader"
)

//export render_app
func render_app(namePtr *C.char, configPtr *C.char, width, height, magnify, maxDuration, timeout C.int, renderGif, silenceOutput C.int) (*C.uchar, C.int) {
	name := C.GoString(namePtr)
	configStr := C.GoString(configPtr)

	var config map[string]string
	err := json.Unmarshal([]byte(configStr), &config)
	if err != nil {
		fmt.Printf("error parsing config: %v\n", err)
		return nil, -1
	}

	result, err := loader.RenderApplet(name, config, int(width), int(height), int(magnify), int(maxDuration), int(timeout), renderGif != 0, silenceOutput != 0)
	if err != nil {
		fmt.Printf("error rendering: %v\n", err)
		return nil, -2
	}
	return (*C.uchar)(C.CBytes(result)), C.int(len(result))
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
