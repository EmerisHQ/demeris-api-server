package ginutils

import (
	"testing"

	"github.com/emerishq/demeris-backend-models/cns"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetValue_Any(t *testing.T) {
	tt := []struct {
		name  string
		value any
	}{
		{
			name:  "int",
			value: 42,
		},
		{
			name:  "string",
			value: "string val",
		},
		{
			name:  "cns.Chain",
			value: cns.Chain{ChainName: "test"},
		},
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(nil)
			c.Set("key", test.value)
			val := GetValue[any](c, "key")
			require.Equal(t, test.value, val)
		})
	}
}

func TestGetValue_Chain(t *testing.T) {
	chain := cns.Chain{ChainName: "test"}

	c, _ := gin.CreateTestContext(nil)
	c.Set("key", chain)

	val := GetValue[cns.Chain](c, "key")
	require.Equal(t, chain.ChainName, val.ChainName)
}
