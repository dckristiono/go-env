package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Parse mengisi struct dari environment variables berdasarkan tag
func (c *Config) Parse(v interface{}) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("expect pointer to struct")
	}

	elem := val.Elem()
	elemType := elem.Type()

	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := elemType.Field(i)

		// Dapatkan tag env
		envTag := fieldType.Tag.Get("env")
		if envTag == "" {
			// Jika tidak ada tag env, gunakan nama field
			envTag = strings.ToUpper(fieldType.Name)
		}

		if !field.CanSet() {
			continue
		}

		prefixedKey := c.prependPrefix(envTag)
		value := os.Getenv(prefixedKey)

		// Dapatkan nilai default dari tag default jika ada
		defaultTag := fieldType.Tag.Get("default")
		if value == "" && defaultTag != "" {
			value = defaultTag
		}

		// Jika masih kosong, lewati
		if value == "" {
			continue
		}

		// Set nilai field berdasarkan tipe
		if err := setFieldValue(field, fieldType, value); err != nil {
			return fmt.Errorf("failed to set field %s: %v", fieldType.Name, err)
		}
	}

	return nil
}

// setFieldValue mengisi nilai field berdasarkan tipe
func setFieldValue(field reflect.Value, fieldType reflect.StructField, value string) error {
	// Isi field berdasarkan tipe
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Periksa apakah tipe Duration
		if fieldType.Type == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value: %v", err)
			}
			field.Set(reflect.ValueOf(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value: %v", err)
			}
			field.SetInt(intVal)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %v", err)
		}
		field.SetUint(uintVal)

	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %v", err)
		}
		field.SetFloat(floatVal)

	case reflect.Bool:
		value = strings.ToLower(value)
		boolVal := value == "true" || value == "1" || value == "yes" || value == "y"
		field.SetBool(boolVal)

	case reflect.Slice:
		if fieldType.Type.Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			// Trim space dari setiap elemen
			slice := reflect.MakeSlice(fieldType.Type, len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("unsupported slice type: %s", fieldType.Type.Elem().Kind())
		}

	case reflect.Map:
		if fieldType.Type.Key().Kind() == reflect.String && fieldType.Type.Elem().Kind() == reflect.String {
			result := reflect.MakeMap(fieldType.Type)
			parts := strings.Split(value, ",")

			for _, part := range parts {
				keyValue := strings.SplitN(part, ":", 2)
				if len(keyValue) == 2 {
					k := strings.TrimSpace(keyValue[0])
					v := strings.TrimSpace(keyValue[1])
					result.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
				}
			}
			field.Set(result)
		} else {
			return fmt.Errorf("unsupported map type: map[%s]%s",
				fieldType.Type.Key().Kind(), fieldType.Type.Elem().Kind())
		}

	default:
		return fmt.Errorf("unsupported type: %s", field.Kind())
	}

	return nil
}

// Parse adalah fungsi level package yang mengisi struct dari environment variables
func Parse(v interface{}) error {
	cfg, err := getDefaultInstance()
	if err != nil {
		return err
	}
	return cfg.Parse(v)
}
