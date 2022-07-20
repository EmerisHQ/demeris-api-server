package options

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestX(t *testing.T) {
	require := require.New(t)
	optionV := Wrap(struct{ SomeKey string }{"somevalue"}, nil)
	j, err := json.Marshal(optionV)
	require.NoError(err)
	require.EqualValues("{\"value\":{\"SomeKey\":\"somevalue\"}}", j)
}

func TestMarshalError(t *testing.T) {
	table := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "string error",
			err:      fmt.Errorf("something bad happened"),
			expected: `{"error":"something bad happened"}`,
		},
		{
			name:     "string error with quotes",
			err:      fmt.Errorf("something \"bad\" happened"),
			expected: `{"error":"something \"bad\" happened"}`,
		},
	}

	for _, test := range table {
		t.Run(test.name, func(t *testing.T) {
			require := require.New(t)
			optionV := Wrap(struct{ SomeKey string }{"somevalue"}, test.err)
			j, err := json.Marshal(optionV)
			require.NoError(err)
			require.Equal(test.expected, string(j))
		})
	}

}
