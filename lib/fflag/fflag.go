// Package fflag provides some utilities to work with feature flags.
//
// This package is intended to be minimal to address a specific use case, if
// we'll start working with more feature flags we might want to move to a more
// robust solution.
//
// A feature flag can be enabled globally, or per request by simply passing the
// name of the feature flag as a query param setting its value to "true".
// Request feature flags takes the precedence over the global ones.
package fflag

import (
	"strings"

	"github.com/gin-gonic/gin"
)

var globalFlags = make(map[string]bool)

// EnableGlobal enables a feature flag globally. Each feature flag can still be
// overridden by query params in each request.
func EnableGlobal(name ...string) {
	for _, n := range name {
		globalFlags[n] = true
	}
}

// Enabled returns true if a particular feature flag is enabled.
func Enabled(c *gin.Context, name string) bool {
	if v, found := c.GetQuery(name); found {
		return strings.ToLower(v) == "true"
	}

	return globalFlags[name]
}
