package services

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

func StructToMap(ins any) map[string]any {
	res := make(map[string]any)
	val := reflect.ValueOf(ins)
	typ := reflect.TypeOf(ins)

	// if pointer then get the element
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
		typ = typ.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		// skip unexported field
		if !field.CanInterface() {
			continue
		}

		tag := fieldType.Tag.Get("json")
		if tag == "" {
			tag = fieldType.Tag.Get("bson")
		}

		if tag == "-" {
			continue
		}

		tagParts := strings.Split(tag, ",")
		keyName := tagParts[0]

		if keyName == "" {
			keyName = strings.ToLower(fieldType.Name)
		}

		if hasModifier("omitempty", tagParts) && field.IsZero() {
			continue
		}

		res[keyName] = field.Interface()

		if hasModifier("string", tagParts) {
			res[keyName] = fmt.Sprintf("%v", field.Interface())
		} else {
			res[keyName] = field.Interface()
		}
	}

	return res
}

func MapToStruct[T any](source map[string]any) (*T, error) {
	instance := new(T)

	val := reflect.ValueOf(instance).Elem()
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		fieldVal := val.Field(i)
		fieldType := typ.Field(i)

		// skip private field
		if !fieldVal.CanSet() {
			continue
		}

		mapKey := strings.ToLower(fieldType.Name)

		tag := fieldType.Tag.Get("bson")
		if tag == "" {
			tag = fieldType.Tag.Get("json")
		}

		if tag == "-" {
			tag = ""
		}

		tagParts := strings.Split(tag, ",")
		keyName := tagParts[0]

		if keyName != "" {
			mapKey = keyName
		}

		mapVal, exists := source[mapKey]

		if !exists {
			continue
		}

		// Get reflection value of the map item
		reflectedMapVal := reflect.ValueOf(mapVal)

		// Safely assign if the types match, or attempt basic conversion if compatible
		if reflectedMapVal.Type().AssignableTo(fieldVal.Type()) {
			fieldVal.Set(reflectedMapVal)
		} else if reflectedMapVal.Type().ConvertibleTo(fieldVal.Type()) {
			// Handles conversions like float64 -> int (common with dynamic values)
			fieldVal.Set(reflectedMapVal.Convert(fieldVal.Type()))
		} else {
			return nil, errors.New("type mismatch for field " + mapKey)
		}
	}

	return instance, nil
}

func GetStructName[T any]() string {
	var ins T
	t := reflect.TypeOf(ins)

	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}

	return t.Name()
}

func hasModifier(part string, parts []string) bool {
	return slices.Contains(parts, part)
}

func FilterFromQueryString(queries []string) map[string]any {
	filter := make(map[string]any)

	for _, query := range queries {
		query = strings.TrimSpace(query)
		parts := strings.SplitN(query, "=", 2)

		if len(parts) < 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) || (strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			filter[key] = val[1 : len(val)-1]
			continue
		}

		if strings.HasPrefix(val, "[") && strings.HasSuffix(val, "]") {
			valSlice := []any{}
			val = val[1 : len(val)-1]

			valParts := strings.Split(val, ",")
			for _, vp := range valParts {
				if (strings.HasPrefix(vp, "\"") && strings.HasSuffix(vp, "\"")) || (strings.HasPrefix(vp, "'") && strings.HasSuffix(vp, "'")) {
					valSlice = append(valSlice, vp[1:len(vp)-1])
					continue
				}

				if intVal, err := strconv.Atoi(vp); err == nil {
					valSlice = append(valSlice, intVal)
					continue
				}

				if floatVal, err := strconv.ParseFloat(vp, 64); err == nil {
					valSlice = append(valSlice, floatVal)
					continue
				}

				if boolVal, err := strconv.ParseBool(vp); err == nil {
					valSlice = append(valSlice, boolVal)
					continue
				}

				valSlice = append(valSlice, vp)
			}
		}

		if intVal, err := strconv.Atoi(val); err == nil {
			filter[key] = intVal
			continue
		}

		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			filter[key] = floatVal
			continue
		}

		if boolVal, err := strconv.ParseBool(val); err == nil {
			filter[key] = boolVal
			continue
		}

		filter[key] = val
	}

	return filter
}

func GetMapValue[K comparable, T any](m map[K]any, key K, defaultVal T) T {
	if rawVal, exists := m[key]; exists {
		if val, ok := rawVal.(T); ok {
			return val
		}
	}
	return defaultVal
}

func SortDirToInt(sortDir string) int {
	dir := -1
	if sortDir == "asc" {
		dir = 1
	}
	return dir
}
