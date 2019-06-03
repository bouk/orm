package rel

import (
	"fmt"
	"reflect"
	"regexp"
)

const assignment = `\s*(\w+)\s*(=|IN)\s*\?\s*`

var assignRe = regexp.MustCompile(`(?:^|,)` + assignment)

// ParseAssignment 'parses' an assignment expression. I'm so sorry
func ParseAssignment(query string, args ...interface{}) ([]Assignment, error) {
	matches := assignRe.FindAllStringSubmatch(query, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("failed to parse %q", query)
	}

	if len(matches) != len(args) {
		return nil, fmt.Errorf("%d arguments passed, %d expected", len(args), len(matches))
	}

	var totalLength int
	exprs := make([]Assignment, len(matches))
	for i, m := range matches {
		field := Field{Name: m[1]}
		value := args[i]
		switch m[2] {
		case "=":
			exprs[i] = Assignment{
				Field: field,
				Value: BindParam{
					Value: value,
				},
			}
		case "IN":
			return nil, fmt.Errorf("failed to parse %q", query)
		}
		totalLength += len(m[0])
	}
	if totalLength != len(query) {
		return nil, fmt.Errorf("failed to parse %q", query)
	}

	return exprs, nil
}

// ParseWhere 'parses' a where expression. I'm so sorry
func ParseWhere(query string, args ...interface{}) ([]Expr, error) {
	matches := assignRe.FindAllStringSubmatch(query, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("failed to parse %q", query)
	}

	if len(matches) != len(args) {
		return nil, fmt.Errorf("%d arguments passed, %d expected", len(args), len(matches))
	}

	var totalLength int
	exprs := make([]Expr, len(matches))
	for i, m := range matches {
		field := Field{Name: m[1]}
		value := args[i]
		switch m[2] {
		case "=":
			exprs[i] = Assignment{
				Field: field,
				Value: BindParam{
					Value: value,
				},
			}
		case "IN":
			v := reflect.ValueOf(value)
			if k := v.Kind(); k != reflect.Slice {
				return nil, fmt.Errorf("expected arg %d to be slice, got %v", i, k)
			}
			l := v.Len()
			right := make([]Expr, l)
			for j := 0; j < l; j++ {
				right[j] = BindParam{
					Value: v.Index(j).Interface(),
				}
			}
			exprs[i] = In{
				Left:  field,
				Right: right,
			}
		}
		totalLength += len(m[0])
	}
	if totalLength != len(query) {
		return nil, fmt.Errorf("failed to parse %q", query)
	}

	return exprs, nil
}
