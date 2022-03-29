package ginutils

import "github.com/gin-gonic/gin"

// GetValue returns a value from the gin context casted to the type parameter T.
func GetValue[T any](c *gin.Context, key string) T {
	v, ok := c.Get(key)
	if !ok {
		panic("key not found in context")
	}
	return v.(T)
}
