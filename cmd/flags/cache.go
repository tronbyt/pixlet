package flags

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/tronbyt/pixlet/runtime"
)

const EnvRedisURL = "PIXLET_REDIS_URL"

func NewCache() *Cache {
	return &Cache{
		RedisURL: os.Getenv(EnvRedisURL),
	}
}

type Cache struct {
	RedisURL string
}

func (c *Cache) Register(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.RedisURL, "redis-url", c.RedisURL, "Redis URL for caching (env "+EnvRedisURL+")")
	_ = cmd.RegisterFlagCompletionFunc("redis-url", cobra.FixedCompletions([]string{"redis://"}, cobra.ShellCompDirectiveNoFileComp))
}

func (c *Cache) Load(ctx context.Context) (runtime.Cache, error) {
	var cache runtime.Cache
	if c.RedisURL != "" {
		var err error
		if cache, err = runtime.NewRedisCache(ctx, c.RedisURL); err != nil {
			return nil, err
		}
	} else {
		cache = runtime.NewInMemoryCache()
	}
	runtime.InitHTTP(cache)
	runtime.InitCache(cache)
	return cache, nil
}
