package apierror

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/sony/sonyflake"
	"go.uber.org/zap"
)

var flake *sonyflake.Sonyflake
var once sync.Once

func init() {
	once.Do(func() {
		flake = sonyflake.NewSonyflake(sonyflake.Settings{})
	})
}

type Error struct {
	ID            string `json:"id"`
	Namespace     string `json:"namespace"`
	StatusCode    int    `json:"-"`
	LowLevelError error  `json:"-"`
	Cause         string `json:"cause"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%s @ %s: %s", e.ID, e.Namespace, e.LowLevelError.Error())
}

func (e Error) Unwrap() error {
	return e.LowLevelError
}

func New(namespace string, cause error, statusCode int) Error {
	id, err := flake.NextID()
	if err != nil {
		panic(fmt.Errorf("cannot create sonyflake, %w", err))
	}

	idstr := strconv.FormatUint(id, 10)

	return Error{
		ID:            idstr,
		StatusCode:    statusCode,
		Namespace:     namespace,
		LowLevelError: cause,
		Cause:         cause.Error(),
	}
}

// WriteError logs and return client-facing errors
func WriteError(logger *zap.SugaredLogger, c *gin.Context, err Error, logMessage string, keyAndValues ...interface{}) {
	_ = c.Error(err)
	LogError(logger, logMessage, keyAndValues...)
}

// LogError logs errors internally without altering the http response
func LogError(logger *zap.SugaredLogger, logMessage string, keyAndValues ...interface{}) {
	if keyAndValues != nil {
		logger.Errorw(
			logMessage,
			keyAndValues...,
		)
	}
}
