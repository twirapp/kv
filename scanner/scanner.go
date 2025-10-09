package kvscanner

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// fieldInfo holds metadata about a struct field for efficient scanning
type fieldInfo struct {
	Index      int
	KVName     string
	LowerName  string
	SnakeName  string
	PascalName string
	CamelName  string
	HasKVTag   bool // Indicates if the field has an explicit kv tag
}

// typeCache stores metadata for each struct type to avoid repeated reflection
var typeCache sync.Map

// Scan scans JSON bytes into a struct, mapping fields by name or kv tag
func Scan(data []byte, dest interface{}) error {
	// Validate destination
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr || destValue.IsNil() {
		return fmt.Errorf("destination must be a non-nil pointer")
	}
	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to a struct")
	}

	// Load or build type metadata
	destType := destValue.Type()
	cacheKey := destType.PkgPath() + "." + destType.Name()
	cached, ok := typeCache.Load(cacheKey)
	var fields []fieldInfo
	if !ok {
		fields = buildFieldCache(destType)
		typeCache.Store(cacheKey, fields)
	} else {
		fields = cached.([]fieldInfo)
	}

	// Unmarshal JSON into a map for flexible field lookup
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(data, &jsonMap); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Map JSON fields to struct fields
	for _, field := range fields {
		var value interface{}
		var found bool

		// If the field has a kv tag, only use KVName
		if field.HasKVTag {
			if v, ok := jsonMap[field.KVName]; ok && field.KVName != "" {
				value = v
				found = true
			}
		} else {
			// Otherwise, try all naming conventions
			for _, name := range []string{
				field.KVName,
				field.LowerName,
				field.SnakeName,
				field.PascalName,
				field.CamelName,
			} {
				if v, ok := jsonMap[name]; ok && name != "" {
					value = v
					found = true
					break
				}
			}
		}

		if !found {
			continue // Skip if no matching field found
		}

		// Set the field value
		fieldValue := destValue.Field(field.Index)
		if !fieldValue.CanSet() {
			return fmt.Errorf("cannot set field %s", field.KVName)
		}

		// Convert JSON value to field type
		if err := setFieldValue(fieldValue, value); err != nil {
			return fmt.Errorf("failed to set field %s: %w", field.KVName, err)
		}
	}

	return nil
}

// buildFieldCache creates metadata for a struct type
func buildFieldCache(t reflect.Type) []fieldInfo {
	var fields []fieldInfo
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		// Get kv tag or use field name
		kvName := f.Tag.Get("kv")
		hasKVTag := kvName != ""
		if kvName == "" {
			kvName = f.Name
		} else {
			// Handle kv tag options (e.g., "name,omitempty")
			if idx := strings.Index(kvName, ","); idx != -1 {
				kvName = kvName[:idx]
			}
		}

		// Generate naming variations
		lowerName := strings.ToLower(f.Name)
		snakeName := toSnakeCase(f.Name)
		pascalName := toPascalCase(f.Name)
		camelName := toCamelCase(f.Name)

		// Check for duplicate field names
		for _, existing := range fields {
			if existing.KVName == kvName || (!hasKVTag && !existing.HasKVTag && (existing.LowerName == lowerName ||
				existing.SnakeName == snakeName ||
				existing.PascalName == pascalName ||
				existing.CamelName == camelName)) {
				panic(fmt.Sprintf("duplicate field name detected for %s in type %s", f.Name, t.Name()))
			}
		}

		fields = append(
			fields, fieldInfo{
				Index:      i,
				KVName:     kvName,
				LowerName:  lowerName,
				SnakeName:  snakeName,
				PascalName: pascalName,
				CamelName:  camelName,
				HasKVTag:   hasKVTag,
			},
		)
	}
	return fields
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && isUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(toLower(r))
	}
	return result.String()
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	var result strings.Builder
	upperNext := true
	for _, r := range s {
		if r == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			result.WriteRune(toUpper(r))
			upperNext = false
		} else {
			result.WriteRune(toLower(r))
		}
	}
	return result.String()
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i == 0 {
			result.WriteRune(toLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// Helper functions for rune case conversion
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func toLower(r rune) rune {
	if isUpper(r) {
		return r + 32
	}
	return r
}

func toUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

// setFieldValue converts and sets a JSON value to a struct field
func setFieldValue(field reflect.Value, value interface{}) error {
	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return nil // Skip null values
	}

	// Handle type conversion
	switch field.Kind() {
	case reflect.String:
		if v.Kind() == reflect.String {
			field.SetString(v.String())
		} else {
			return fmt.Errorf("expected string, got %v", v.Kind())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Kind() == reflect.Float64 {
			n := int64(v.Float())
			if field.OverflowInt(n) {
				return fmt.Errorf("integer overflow for %v", n)
			}
			field.SetInt(n)
		} else {
			return fmt.Errorf("expected number, got %v", v.Kind())
		}
	case reflect.Float32, reflect.Float64:
		if v.Kind() == reflect.Float64 {
			field.SetFloat(v.Float())
		} else {
			return fmt.Errorf("expected number, got %v", v.Kind())
		}
	case reflect.Bool:
		if v.Kind() == reflect.Bool {
			field.SetBool(v.Bool())
		} else {
			return fmt.Errorf("expected bool, got %v", v.Kind())
		}
	case reflect.Slice:
		if v.Kind() == reflect.Slice {
			slice := reflect.MakeSlice(field.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				if err := setFieldValue(slice.Index(i), v.Index(i).Interface()); err != nil {
					return err
				}
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("expected slice, got %v", v.Kind())
		}
	case reflect.Struct:
		// Handle nested structs recursively
		if v.Kind() == reflect.Map {
			data, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("failed to marshal nested struct: %w", err)
			}
			newStruct := reflect.New(field.Type())
			if err := Scan(data, newStruct.Interface()); err != nil {
				return err
			}
			field.Set(newStruct.Elem())
		} else {
			return fmt.Errorf("expected object, got %v", v.Kind())
		}
	default:
		return fmt.Errorf("unsupported field type: %v", field.Kind())
	}
	return nil
}
