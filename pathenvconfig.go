package pathenvconfig

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/dlclark/regexp2"
)

const (
	EnvironmentVariableFileSuffix = "_FILE"
)

var ErrInvalidSpecification = errors.New("specification must be a struct pointer")

var re = regexp2.MustCompile(`([A-Z]+(?![a-z]))|([A-Z]?[a-z0-9]+)`, 0)

func Process(prefix string, spec interface{}) error {
	if len(prefix) > 0 && !strings.HasSuffix("_", prefix) {
		prefix += "_"
	}

	specValue := reflect.ValueOf(spec)
	if specValue.Kind() != reflect.Ptr {
		return ErrInvalidSpecification
	}
	specValue = specValue.Elem()
	if specValue.Kind() != reflect.Struct {
		return ErrInvalidSpecification
	}

	structType := specValue.Type()

	for i := 0; i < structType.NumField(); i++ {

		field := specValue.Field(i)
		if !field.CanSet() {
			continue
		}

		fieldType := structType.Field(i)

		var defaultValue *string

		required, _ := strconv.ParseBool(fieldType.Tag.Get("required"))
		if defaultTagValue, ok := fieldType.Tag.Lookup("default"); ok {
			defaultValue = &defaultTagValue
		}

		variableName := fieldNameToEnvironmentVar(prefix, fieldType.Name)

		pointerToField := specValue.Field(i).Addr().Interface()

		if err := getEnvironmentValue(pointerToField, variableName, required, defaultValue); err != nil {
			return err
		}
	}

	return nil
}

func fieldNameToEnvironmentVar(prefix, fieldName string) string {
	matches, err := getAllMatches(re, fieldName)
	if err != nil {
		log.Panicf("Failed to to turn field name '%s' in an environment variable: %v", fieldName, err)
	}

	return prefix + strings.ToUpper(strings.Join(matches, "_"))
}

func getEnvironmentValue(fieldPointer interface{}, variableName string, required bool, defaultValue *string) error {
	originalVariableName := variableName
	val, found := os.LookupEnv(variableName)
	if !found {
		variableName += EnvironmentVariableFileSuffix
		path, found := os.LookupEnv(variableName)

		if found {
			bytes, err := os.ReadFile(path)
			if err != nil {
				log.Panicf("Unable to read file %s for environment variable %s", path, variableName)
			}

			val = string(bytes)
		} else {
			if defaultValue == nil {
				if required {
					return fmt.Errorf("neither '%s' nor '%s' were provided as environment variables. One of them is required", originalVariableName, variableName)
				}
				return nil
			}

			val = *defaultValue
		}
	}

	_, err := fmt.Sscan(val, fieldPointer)
	if err != nil {
		return fmt.Errorf("unable to convert value for environment variable '%s' to target type %t", variableName, reflect.TypeOf(fieldPointer))
	}

	return nil
}

func getAllMatches(re *regexp2.Regexp, s string) ([]string, error) {
	var matches []string
	m, err := re.FindStringMatch(s)
	if err != nil {
		return nil, err
	}

	for m != nil {
		matches = append(matches, m.String())
		m, err = re.FindNextMatch(m)
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}
