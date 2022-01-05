package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

const (
	CorrelationIDName    = "correlation_id"
	IntCorrelationIDName = "int_correlation_id"
)

// CorrelationIDMiddleware adds correlationID if it's not specified in HTTP request
func CorrelationIDMiddleware(l *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		addCorrelationID(c, l)
	}
}

func addCorrelationID(c *gin.Context, l *zap.SugaredLogger) {
	ctx := c.Request.Context()

	correlationID := c.Request.Header.Get("X-Correlation-id")

	if correlationID != "" {
		ctx = context.WithValue(ctx, CorrelationIDName, correlationID)
		c.Writer.Header().Set("X-Correlation-Id", correlationID)
		l = l.With(CorrelationIDName, correlationID)
	}

	id, _ := uuid.NewV4()

	ctx = context.WithValue(ctx, IntCorrelationIDName, id.String())
	l = l.With(IntCorrelationIDName, id)

	c.Set("logger", l)

	c.Request = c.Request.WithContext(ctx)

	c.Next()
}

func getLoggerFromContext(c *gin.Context) *zap.SugaredLogger {
	value, ok := c.Get("logger")
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, "logger does not exists in context")
	}

	l, ok := value.(*zap.SugaredLogger)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, "invalid logger format in context")
	}

	return l
}
