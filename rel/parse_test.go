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
		exprs []Assignment
		err   error
	}{
		{
			name:  "single",
			query: "id = ?",
			args:  []interface{}{1},
			exprs: []Assignment{
				{
					Field: "id",
					Value: BindParam{1},
				},
			},
		},
		{
			name:  "two",
			query: "a = ?, b = ?",
			args:  []interface{}{1, 2},
			exprs: []Assignment{
				{
					Field: "a",
					Value: BindParam{1},
				},
				{
					Field: "b",
					Value: BindParam{2},
				},
			},
		},
		{
			name:  "many",
			query: "a = ?, b = ?, c = ?",
			args:  []interface{}{1, 2, 3},
			exprs: []Assignment{
				{
					Field: "a",
					Value: BindParam{1},
				},
				{
					Field: "b",
					Value: BindParam{2},
				},
				{
					Field: "c",
					Value: BindParam{3},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := ParseAssignment(tt.query, tt.args...)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.exprs, exprs)
		})
	}
}
