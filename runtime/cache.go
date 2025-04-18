package runtime

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const DefaultExpirationSeconds = 60

type Cache interface {
	Set(thread *starlark.Thread, key string, value []byte, ttl int64) error
	Get(thread *starlark.Thread, key string) ([]byte, bool, error)
}

type InMemoryCacheRecord struct {
	data       []byte
	expiration time.Time
}

type InMemoryCache struct {
	records map[string]*InMemoryCacheRecord
	mutex   sync.RWMutex
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{records: map[string]*InMemoryCacheRecord{}}
}

func (c *InMemoryCache) Get(_ *starlark.Thread, key string) (value []byte, found bool, err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	r, found := c.records[key]

	if !found {
		return nil, false, nil
	}

	if time.Now().After(r.expiration) {
		return nil, false, nil
	}

	return r.data, true, nil
}

func (c *InMemoryCache) Set(_ *starlark.Thread, key string, value []byte, ttl int64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.records[key] = &InMemoryCacheRecord{
		data:       value,
		expiration: time.Now().Add(time.Duration(ttl) * time.Second),
	}

	return nil
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(url string) *RedisCache {
	opts, err := redis.ParseURL(url)
	if err != nil {
		panic(err)
	}

	return &RedisCache{
		client: redis.NewClient(opts),
	}
}

func (c *RedisCache) Get(_ *starlark.Thread, key string) (value []byte, found bool, err error) {
	ctx := context.Background()
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, false, nil
	} else if err != nil {
		return nil, false, err
	} else {
		return val, true, nil
	}
}

func (c *RedisCache) Set(_ *starlark.Thread, key string, value []byte, ttl int64) error {
	ctx := context.Background()
	return c.client.Set(ctx, key, value, time.Duration(ttl)*time.Second).Err()
}

var (
	cacheOnce   sync.Once
	cacheModule starlark.StringDict
	cache       Cache
)

func InitCache(c Cache) {
	cache = c
}

func LoadCacheModule() (starlark.StringDict, error) {
	cacheOnce.Do(func() {
		cacheModule = starlark.StringDict{
			"cache": &starlarkstruct.Module{
				Name: "cache",
				Members: starlark.StringDict{
					"get": starlark.NewBuiltin("get", cacheGet),
					"set": starlark.NewBuiltin("set", cacheSet),
				},
			},
		}
	})

	return cacheModule, nil
}

func scopedCacheKey(thread *starlark.Thread, key starlark.String) string {
	return fmt.Sprintf("pixlet:%s:%s", thread.Name, key.GoString())
}

func cacheGet(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var key starlark.String

	if err := starlark.UnpackArgs(
		"get",
		args, kwargs,
		"key", &key,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for cache.get: %v", err)
	}

	cacheKey := scopedCacheKey(thread, key)

	if cache == nil {
		// no cache configured
		return starlark.None, nil
	}

	val, found, err := cache.Get(thread, cacheKey)

	if err != nil {
		// don't fail just because cache is misbehaving
		log.Printf("getting %s from cache: %v", cacheKey, err)
		return starlark.None, nil
	}

	if !found {
		return starlark.None, nil
	}

	return starlark.String(val), nil
}

func cacheSet(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		key starlark.String
		val starlark.String
		ttl starlark.Int
	)

	if err := starlark.UnpackArgs(
		"set",
		args, kwargs,
		"key", &key,
		"value", &val,
		"ttl_seconds?", &ttl,
	); err != nil {
		return nil, fmt.Errorf("unpacking arguments for cache.set: %v", err)
	}

	cacheKey := scopedCacheKey(thread, key)

	ttl64, ok := ttl.Int64()
	if !ok {
		return nil, fmt.Errorf("ttl_seconds must be valid integer (not %s)", ttl.String())
	}

	if ttl64 < 0 {
		return nil, fmt.Errorf("ttl_seconds cannot be negative")
	}

	if ttl64 == 0 {
		ttl64 = DefaultExpirationSeconds
	}

	if cache == nil {
		// no cache configured
		return starlark.None, nil
	}

	err := cache.Set(thread, cacheKey, []byte(val.GoString()), ttl64)
	if err != nil {
		log.Printf("setting %s in cache: %v", cacheKey, err)
	}

	return starlark.None, nil
}
