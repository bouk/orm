package rel

import (
	"fmt"
	"regexp"
)

const assignment = `\s*(\w+)\s*=\s*\?\s*`

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
		exprs[i] = Assignment{
			Field: m[1],
			Value: BindParam{
				Value: args[i],
			},
		}
		totalLength += len(m[0])
	}
	if totalLength != len(query) {
		return nil, fmt.Errorf("failed to parse %q", query)
	}

	return exprs, nil
}
