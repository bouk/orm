package rel

import (
	"errors"
	"fmt"
	"regexp"
)

const assignment = `\s*(\w+)\s*=\s*\?\s*`

var assignRe = regexp.MustCompile(`^` + assignment + `(?:,` + assignment + `)*$`)

// ParseAssignment 'parses' an assignment expression. I'm so sorry
func ParseAssignment(query string, args ...interface{}) ([]Expr, error) {
	matches := assignRe.FindStringSubmatch(query)
	if len(matches) == 0 {
		return nil, errors.New("failed to parse")
	}

	if matches[len(matches)-1] == "" {
		matches = matches[:len(matches)-1]
	}

	if len(matches)-1 != len(args) {
		return nil, fmt.Errorf("%d arguments passed, %d expected", len(args), len(matches)-1)
	}

	exprs := make([]Expr, len(matches)-1)
	for i, m := range matches[1:] {
		exprs[i] = &Equality{
			Field: m,
			Expr: &BindParam{
				Value: args[i],
			},
		}
	}
	return exprs, nil
}
