package cache

import (
	"context"
	"fmt"
	"time"

	gocache "github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/marshaler"
	cache_metrics "github.com/eko/gocache/lib/v4/metrics"
	"github.com/eko/gocache/lib/v4/store"
	redis_store "github.com/eko/gocache/store/redis/v4"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/db"
	"github.com/nbrglm/nexeres/opts"
	"github.com/redis/go-redis/v9"
)

var cached *marshaler.Marshaler

func InitCache() error {
	redisPassword := ""
	if config.C.Stores.Redis.Password != nil {
		redisPassword = *config.C.Stores.Redis.Password
	}
	redisOpts := redis.Options{
		Addr:     config.C.Stores.Redis.Address,
		Password: redisPassword,
		DB:       config.C.Stores.Redis.DB,
	}
	redisClient := redis.NewClient(&redisOpts)

	// Initialize the Redis cache
	redisStore := redis_store.NewRedis(redisClient)

	metrics := cache_metrics.NewPrometheus(fmt.Sprintf("%s_cache", opts.Name))

	cacheManager := gocache.NewMetric(metrics, redisStore)
	cached = marshaler.New(cacheManager)
	return nil
}

type FlowType string

var (
	FlowTypeLogin          FlowType = "login"           // For Login Flow
	FlowTypeChangePassword FlowType = "change-password" // For Change Password Flow (user already logged in)
	FlowTypeSSO            FlowType = "sso"             // TODO: Implement SSO flow
)

type FlowData struct {
	ID          string    `json:"id"`
	Type        FlowType  `json:"type"`
	UserID      string    `json:"userId"`
	Email       string    `json:"email"`
	Orgs        []db.Org  `json:"orgs,omitempty"`
	MFARequired bool      `json:"mfaRequired"`
	MFAVerified bool      `json:"mfaVerified"`
	ReturnTo    *string   `json:"returnTo,omitempty"` // URL to redirect after flow completion
	CreatedAt   time.Time `json:"createdAt"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

func StoreFlow(ctx context.Context, flow FlowData) error {
	exp := time.Until(flow.ExpiresAt)
	return cached.Set(ctx, flow.ID, flow, store.WithExpiration(exp))
}

var ErrKeyNotFound = fmt.Errorf("flow not found")

// GetFlow retrieves a flow by its ID from the cache.
//
// IMP: DO NOT RETURN nil for error if flow is not found, return a specific error instead.
func GetFlow(ctx context.Context, flowID string) (*FlowData, error) {
	if flow, err := cached.Get(ctx, flowID, new(FlowData)); err != nil {
		if err.Error() == store.NOT_FOUND_ERR {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get flow: %w", err)
	} else {
		if f, ok := flow.(*FlowData); !ok || f == nil {
			return nil, fmt.Errorf("invalid flow data stored")
		} else {
			return f, nil
		}
	}
}

type AdminLoginFlowData struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func StoreAdminLoginFlow(ctx context.Context, flow AdminLoginFlowData) error {
	exp := time.Until(flow.ExpiresAt)
	return cached.Set(ctx, fmt.Sprintf("nexeres_admin_login_flow:%s", flow.ID), flow, store.WithExpiration(exp))
}

// GetAdminLoginFlow retrieves an admin login flow by its ID from the cache.
//
// IMP: DO NOT RETURN nil for error if flow is not found, return a specific error instead.
func GetAdminLoginFlow(ctx context.Context, flowID string) (*AdminLoginFlowData, error) {
	if flow, err := cached.Get(ctx, fmt.Sprintf("nexeres_admin_login_flow:%s", flowID), new(AdminLoginFlowData)); err != nil {
		if err.Error() == store.NOT_FOUND_ERR {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get admin login flow: %w", err)
	} else {
		if f, ok := flow.(*AdminLoginFlowData); !ok || f == nil {
			return nil, fmt.Errorf("invalid admin login flow data stored")
		} else {
			return f, nil
		}
	}
}

func DeleteAdminLoginFlow(ctx context.Context, flowID string) error {
	if err := cached.Delete(ctx, fmt.Sprintf("nexeres_admin_login_flow:%s", flowID)); err != nil {
		if err.Error() == store.NOT_FOUND_ERR {
			return ErrKeyNotFound
		}
		return fmt.Errorf("failed to delete admin login flow: %w", err)
	}
	return nil
}

type AdminSessionData struct {
	Email string `json:"email"`
	// The token is the ID/Key for the session
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func StoreAdminSession(ctx context.Context, session AdminSessionData) error {
	exp := time.Until(session.ExpiresAt)
	return cached.Set(ctx, fmt.Sprintf("nexeres_admin_session:%s", session.Token), session, store.WithExpiration(exp))
}

// GetAdminSession retrieves an admin session by its token from the cache.
//
// IMP: DO NOT RETURN nil for error if session is not found, return a specific error instead.
func GetAdminSession(ctx context.Context, token string) (*AdminSessionData, error) {
	if session, err := cached.Get(ctx, fmt.Sprintf("nexeres_admin_session:%s", token), new(AdminSessionData)); err != nil {
		if err.Error() == store.NOT_FOUND_ERR {
			return nil, ErrKeyNotFound
		}
		return nil, fmt.Errorf("failed to get admin session: %w", err)
	} else {
		if s, ok := session.(*AdminSessionData); !ok || s == nil {
			return nil, fmt.Errorf("invalid admin session data stored")
		} else {
			return s, nil
		}
	}
}
