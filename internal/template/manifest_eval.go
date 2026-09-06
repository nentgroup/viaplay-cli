package template

import (
	"fmt"
	"strconv"
	"strings"
)

// EvalCondition evaluates a small manifest condition language.
func EvalCondition(expr string, vars *Variables) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true, nil
	}
	if strings.Contains(expr, "||") {
		parts := strings.Split(expr, "||")
		for _, part := range parts {
			ok, err := EvalCondition(part, vars)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
		return false, nil
	}
	if strings.Contains(expr, "&&") {
		parts := strings.Split(expr, "&&")
		for _, part := range parts {
			ok, err := EvalCondition(part, vars)
			if err != nil || !ok {
				return ok, err
			}
		}
		return true, nil
	}
	if strings.HasPrefix(expr, "!") {
		ok, err := EvalCondition(strings.TrimSpace(strings.TrimPrefix(expr, "!")), vars)
		return !ok, err
	}
	for _, op := range []string{"==", "!="} {
		if parts := strings.Split(expr, op); len(parts) == 2 {
			left := strings.TrimSpace(parts[0])
			right := strings.TrimSpace(parts[1])
			lv := lookupConditionValue(left, vars)
			rv := parseConditionLiteral(right)
			if op == "==" {
				return fmt.Sprint(lv) == fmt.Sprint(rv), nil
			}
			return fmt.Sprint(lv) != fmt.Sprint(rv), nil
		}
	}
	return truthy(lookupConditionValue(expr, vars)), nil
}

func lookupConditionValue(path string, vars *Variables) any {
	if vars == nil {
		return nil
	}
	key := strings.TrimSpace(path)
	switch key {
	case "true":
		return true
	case "false":
		return false
	default:
		if vars.Features == nil {
			return nil
		}
		if v, ok := vars.Features[key]; ok {
			return v
		}
		return nil
	}
}

func parseConditionLiteral(v string) any {
	v = strings.TrimSpace(strings.Trim(v, `"'`))
	if b, err := strconv.ParseBool(v); err == nil {
		return b
	}
	return v
}

func truthy(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != "" && t != "false" && t != "0"
	case nil:
		return false
	default:
		return true
	}
}
