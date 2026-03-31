package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/httprate"
	httprate_redis "github.com/go-chi/httprate-redis"

	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/opts"
)

func RateLimitMiddleware() func(http.Handler) http.Handler {
	redisPassword := ""
	if config.C.Stores.Redis.Password != nil {
		redisPassword = *config.C.Stores.Redis.Password
	}
	prefix := fmt.Sprintf("%s_rate_limit", opts.Name)
	addrParts := strings.Split(config.C.Stores.Redis.Address, ":")
	port, err := strconv.Atoi(addrParts[1])
	if err != nil {
		port = 6379
	}
	unit := time.Second
	switch config.C.Security.RateLimiting.Duration {
	case "s":
		unit = time.Second
	case "m":
		unit = time.Minute
	case "h":
		unit = time.Hour
	}

	return httprate.Limit(
		config.C.Security.RateLimiting.Rate,
		unit,
		httprate.WithKeyFuncs(rateLimitKeyFunc),
		httprate_redis.WithRedisLimitCounter(&httprate_redis.Config{
			Host:      addrParts[0],
			Port:      uint16(port),
			PrefixKey: prefix,
			DBIndex:   config.C.Stores.Redis.DB,
			Password:  redisPassword,
		}))
}

func rateLimitKeyFunc(r *http.Request) (string, error) {
	return "", nil
}

// RateLimitMiddleware returns a middleware that applies rate limiting
// to API endpoints based on the configured rate limit.
