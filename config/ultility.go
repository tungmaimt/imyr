package config

import (
	"strconv"
	"strings"
)

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
