package icontext

import (
	"context"

	"github.com/portalenergy/pe-api-admin/app/models"
	"github.com/rs/xid"
	log "github.com/sirupsen/logrus"
)

type key string

const (
	UserContext         = key("userContext")
	LoggerContextKey    = key("loggerContextKey")
	RequestIDContextKey = key("requestIDContextKey")
)

func GetContext() context.Context {
	ctx := context.Background()
	guid := xid.New()
	requestID := guid.String()
	requestLogger := log.WithFields(log.Fields{"request_id": requestID})
	ctx = context.WithValue(ctx, LoggerContextKey, requestLogger)
	ctx = context.WithValue(ctx, RequestIDContextKey, requestID)
	return ctx
}

func GetRequestID(ctx context.Context) (*string, bool) {
	u, ok := ctx.Value(RequestIDContextKey).(*string)
	return u, ok
}

//todo add get user method
func GetUser(ctx context.Context) (*models.User, bool) {
	u, ok := ctx.Value(UserContext).(*models.User)
	return u, ok
}

// GetLogger - return logger instance from context if it exists.
func GetLogger(ctx context.Context) (*log.Entry, bool) {
	u, ok := ctx.Value(LoggerContextKey).(*log.Entry)
	return u, ok
}
