package engine

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

var dynamicInterfaceReservedArgs = map[string]struct{}{
	"node_id":                  {},
	"external_interface":       {},
	"delivery_mode":            {},
	"mode":                     {},
	"primary_transport":        {},
	"integration":              {},
	"transport":                {},
	"timeout_ms":               {},
	"consumer":                 {},
	"max_attempts":             {},
	"callback_post_process":    {},
	"callback_memory_level":    {},
	"callback_memory_template": {},
}

func validateDynamicActionArgs(args map[string]any, schema map[string]any) error {
	if len(schema) == 0 {
		return nil
	}
	return validateSchemaValue(args, schema, "args")
}

func validateSchemaValue(value any, schema map[string]any, path string) error {
	if len(schema) == 0 {
		return nil
	}
	if err := validateSchemaType(value, schema, path); err != nil {
		return err
	}
	if err := validateSchemaEnum(value, schema, path); err != nil {
		return err
	}
	if objectValue, ok := value.(map[string]any); ok {
		if err := validateSchemaObject(objectValue, schema, path); err != nil {
			return err
		}
	}
	if arrayValue, ok := value.([]any); ok {
		if err := validateSchemaArray(arrayValue, schema, path); err != nil {
			return err
		}
	}
	return nil
}

func validateSchemaType(value any, schema map[string]any, path string) error {
	rawType, _ := schema["type"].(string)
	schemaType := strings.TrimSpace(rawType)
	if schemaType == "" {
		return nil
	}
	if typeMatchesSchema(value, schemaType) {
		return nil
	}
	return fmt.Errorf("%s must be %s", path, schemaType)
}

func validateSchemaEnum(value any, schema map[string]any, path string) error {
	rawEnum, ok := schema["enum"]
	if !ok {
		return nil
	}
	items, ok := rawEnum.([]any)
	if !ok || len(items) == 0 {
		return nil
	}
	for _, item := range items {
		if schemaValuesEqual(value, item) {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of the allowed enum values", path)
}

func validateSchemaObject(value map[string]any, schema map[string]any, path string) error {
	properties := schemaProperties(schema)
	for _, key := range schemaRequiredKeys(schema) {
		if _, reserved := dynamicInterfaceReservedArgs[key]; reserved {
			continue
		}
		if _, ok := value[key]; !ok {
			return fmt.Errorf("%s.%s is required", path, key)
		}
	}
	allowAdditional := true
	if raw, ok := schema["additionalProperties"].(bool); ok {
		allowAdditional = raw
	}
	if !allowAdditional {
		for key := range value {
			if _, reserved := dynamicInterfaceReservedArgs[key]; reserved {
				continue
			}
			if _, ok := properties[key]; !ok {
				return fmt.Errorf("%s.%s is not allowed", path, key)
			}
		}
	}
	for key, childSchema := range properties {
		if _, reserved := dynamicInterfaceReservedArgs[key]; reserved {
			continue
		}
		childValue, ok := value[key]
		if !ok {
			continue
		}
		if err := validateSchemaValue(childValue, childSchema, path+"."+key); err != nil {
			return err
		}
	}
	return nil
}

func validateSchemaArray(value []any, schema map[string]any, path string) error {
	if minItems, ok := schemaInt(schema["minItems"]); ok && len(value) < minItems {
		return fmt.Errorf("%s must contain at least %d items", path, minItems)
	}
	if maxItems, ok := schemaInt(schema["maxItems"]); ok && len(value) > maxItems {
		return fmt.Errorf("%s must contain at most %d items", path, maxItems)
	}
	itemSchema, ok := schema["items"].(map[string]any)
	if !ok || len(itemSchema) == 0 {
		return nil
	}
	for idx, item := range value {
		if err := validateSchemaValue(item, itemSchema, fmt.Sprintf("%s[%d]", path, idx)); err != nil {
			return err
		}
	}
	return nil
}

func schemaProperties(schema map[string]any) map[string]map[string]any {
	result := map[string]map[string]any{}
	rawProperties, ok := schema["properties"].(map[string]any)
	if !ok {
		return result
	}
	for key, raw := range rawProperties {
		child, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		result[key] = child
	}
	return result
}

func schemaRequiredKeys(schema map[string]any) []string {
	rawRequired, ok := schema["required"]
	if !ok {
		return nil
	}
	switch items := rawRequired.(type) {
	case []string:
		return items
	case []any:
		result := make([]string, 0, len(items))
		for _, item := range items {
			text, ok := item.(string)
			if !ok || strings.TrimSpace(text) == "" {
				continue
			}
			result = append(result, text)
		}
		return result
	default:
		return nil
	}
}

func schemaInt(value any) (int, bool) {
	switch num := value.(type) {
	case int:
		return num, true
	case int64:
		return int(num), true
	case float64:
		if math.Trunc(num) != num {
			return 0, false
		}
		return int(num), true
	default:
		return 0, false
	}
}

func schemaValuesEqual(left, right any) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}
	leftNum, leftOK := normalizeSchemaNumber(left)
	rightNum, rightOK := normalizeSchemaNumber(right)
	if leftOK && rightOK {
		return leftNum == rightNum
	}
	return false
}

func normalizeSchemaNumber(value any) (float64, bool) {
	switch num := value.(type) {
	case int:
		return float64(num), true
	case int64:
		return float64(num), true
	case float64:
		return num, true
	default:
		return 0, false
	}
}
