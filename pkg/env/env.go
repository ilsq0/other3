package env

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
)

func LoadConfig(cfg any) error {
	v := reflect.ValueOf(cfg)

	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("LoadConfig only accepts a pointer to a struct")
	}

	// Get the reflect.Value of the struct (dereference the pointer)
	v = v.Elem()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		if fieldType.IsExported() {
			envValue, err := getEnv("env" + fieldType.Name)
			if err != nil {
				return err
			}
			if err = setFieldFromString(field, envValue); err != nil {
				return fmt.Errorf("error setting field %s: %v", fieldType.Name, err)
			}
		}
	}

	return nil
}

func setFieldFromString(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue, err := parseInteger(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(intValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintValue, err := parseUnsignedInteger(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(uintValue)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(boolValue)
	default:
		return fmt.Errorf("unsupported field type: %v", field.Type())
	}
	return nil
}

// parseInteger parses a string into a signed integer based on the provided bit size.
func parseInteger(value string, bitSize int) (int64, error) {
	return strconv.ParseInt(value, 10, bitSize)
}

// parseUnsignedInteger parses a string into an unsigned integer based on the provided bit size.
func parseUnsignedInteger(value string, bitSize int) (uint64, error) {
	return strconv.ParseUint(value, 10, bitSize)
}

func getEnv(key string) (string, error) {
	val, exists := os.LookupEnv(key)
	if !exists {
		return "", fmt.Errorf("environment variable %s missing", key)
	}
	return val, nil
}