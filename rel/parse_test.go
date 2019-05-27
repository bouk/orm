package rel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAssignment(t *testing.T) {
	for _, tt := range []struct {
		name  string
		query string
		args  []interface{}
		err   error
	}{
		{
			name:  "single",
			query: "id = ?",
			args:  []interface{}{1},
		},
		{
			name:  "multiple",
			query: "a = ?, b = ?",
			args:  []interface{}{1, 2},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseAssignment(tt.query, tt.args...)
			require.Equal(t, tt.err, err)
		})
	}
}
