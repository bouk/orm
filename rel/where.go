package rel

import (
	"fmt"
	"reflect"
)

// UnpackWhere creates a list of expression from a Where call
func UnpackWhere(value interface{}, args ...interface{}) ([]Expr, error) {
	switch value := value.(type) {
	case string:
		return []Expr{
			Literal{
				Text:   value,
				Params: args,
			},
		}, nil
	case Expr:
		if len(args) != 0 {
			return nil, fmt.Errorf("unexpected args with expr")
		}
		return []Expr{value}, nil
	default:
		v := reflect.ValueOf(value)
		if v.Kind() != reflect.Map || v.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("invalid %T", value)
		}

		if len(args) != 0 {
			return nil, fmt.Errorf("unexpected args with map")
		}

		exprs := make([]Expr, 0, v.Len())
		for i := v.MapRange(); i.Next(); {
			exprs = append(exprs, Equality{
				Field: Field{i.Key().String()},
				Value: BindParam{i.Value().Interface()},
			})
		}

		return exprs, nil
	}
}
