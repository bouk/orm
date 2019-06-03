package rel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseWhere(t *testing.T) {
	for _, tt := range []struct {
		name  string
		query string
		args  []interface{}
		exprs []Expr
		err   error
	}{
		{
			name:  "single",
			query: "id = ?",
			args:  []interface{}{1},
			exprs: []Expr{
				Assignment{
					Field: Field{"id"},
					Value: BindParam{1},
				},
			},
		},
		{
			name:  "two",
			query: "a = ?, b = ?",
			args:  []interface{}{1, 2},
			exprs: []Expr{
				Assignment{
					Field: Field{"a"},
					Value: BindParam{1},
				},
				Assignment{
					Field: Field{"b"},
					Value: BindParam{2},
				},
			},
		},
		{
			name:  "many",
			query: "a = ?, b = ?, c = ?",
			args:  []interface{}{1, 2, 3},
			exprs: []Expr{
				Assignment{
					Field: Field{"a"},
					Value: BindParam{1},
				},
				Assignment{
					Field: Field{"b"},
					Value: BindParam{2},
				},
				Assignment{
					Field: Field{"c"},
					Value: BindParam{3},
				},
			},
		},
		{
			name:  "IN",
			query: "a IN ?, b = ?, c IN ?",
			args:  []interface{}{[]int{1, 2, 3}, 4, []interface{}{"hi", "dog"}},
			exprs: []Expr{
				In{
					Left: Field{"a"},
					Right: []Expr{
						BindParam{1},
						BindParam{2},
						BindParam{3},
					},
				},
				Assignment{
					Field: Field{"b"},
					Value: BindParam{4},
				},
				In{
					Left: Field{"c"},
					Right: []Expr{
						BindParam{"hi"},
						BindParam{"dog"},
					},
				},
			},
		},
		{
			name:  "IN needs slice",
			query: "a IN ?",
			args:  []interface{}{1},
			err:   errors.New("expected arg 0 to be slice, got int"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := ParseWhere(tt.query, tt.args...)
			require.Equal(t, tt.err, err)
			require.Equal(t, tt.exprs, exprs)
		})
	}
}
