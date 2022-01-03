package router

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

const (
	correlationIDName    = "correlation_id"
	intCorrelationIDName = "int_correlation_id"
)

//CorrelationIDMiddleware adds correlationID if it's not specified in HTTP request
func CorrelationIDMiddleware(l *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		addCorrelationID(c, l)
	}
}

func addCorrelationID(c *gin.Context, l *zap.SugaredLogger) {
	ctx := c.Request.Context()

	corralationID := c.Request.Header.Get("X-Correlation-id")

	if corralationID != "" {
		ctx = context.WithValue(ctx, correlationIDName, corralationID)
		c.Request.Response.Header.Set("X-Correlation-Id", corralationID)
		l = l.With(correlationIDName, corralationID)
	}

	id, _ := uuid.NewV4()

	ctx = context.WithValue(ctx, intCorrelationIDName, id.String())
	l = l.With(intCorrelationIDName, id)

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
