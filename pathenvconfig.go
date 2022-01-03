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
	_, err := ProcessImpl(prefix, spec)
	return err
}

func ProcessImpl(prefix string, spec interface{}) (changed bool, err error) {
	if len(prefix) > 0 && !strings.HasSuffix("_", prefix) {
		prefix += "_"
	}

	specValue := reflect.ValueOf(spec)
	if specValue.Kind() != reflect.Ptr {
		return changed, ErrInvalidSpecification
	}
	specValue = specValue.Elem()
	if specValue.Kind() != reflect.Struct {
		return changed, ErrInvalidSpecification
	}

	structType := specValue.Type()

	for i := 0; i < structType.NumField(); i++ {

		fieldValue := specValue.Field(i)
		if !fieldValue.CanSet() {
			continue
		}

		fieldType := structType.Field(i)

		var defaultValue *string

		required, _ := strconv.ParseBool(fieldType.Tag.Get("required"))
		if defaultTagValue, ok := fieldType.Tag.Lookup("default"); ok {
			defaultValue = &defaultTagValue
		}

		variableName := fieldNameToEnvironmentVar(prefix, fieldType.Name)

		pointerToField := fieldValue.Addr().Interface()

		if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			wasNil := fieldValue.IsNil()
			var structPtr interface{}
			if wasNil {
				// allocate a value
				value := reflect.New(fieldType.Type.Elem())
				fieldValue.Set(value)
				structPtr = value.Interface()
			} else {
				structPtr = fieldValue.Interface()
			}

			valueSet, err := ProcessImpl(variableName, structPtr)
			if err != nil {
				return changed, err
			}
			if !valueSet && wasNil {
				// revert back to nil
				fieldValue.Set(reflect.Zero(fieldType.Type))
			}
			changed = changed || valueSet
		} else if fieldType.Type.Kind() == reflect.Struct {
			valueSet, err := ProcessImpl(variableName, pointerToField)
			if err != nil {
				return changed, err
			}
			changed = changed || valueSet
		} else {
			valueSet, err := assignFromEnvironmentValue(pointerToField, variableName, required, defaultValue)
			if err != nil {
				return changed, err
			}
			changed = changed || valueSet
		}
	}

	return changed, nil
}

func fieldNameToEnvironmentVar(prefix, fieldName string) string {
	matches, err := getAllMatches(re, fieldName)
	if err != nil {
		log.Panicf("Failed to to turn field name '%s' in an environment variable: %v", fieldName, err)
	}

	return prefix + strings.ToUpper(strings.Join(matches, "_"))
}

func assignFromEnvironmentValue(fieldPointer interface{}, variableName string, required bool, defaultValue *string) (changed bool, err error) {
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
					return false, fmt.Errorf("neither '%s' nor '%s' were provided as environment variables. One of them is required", originalVariableName, variableName)
				}
				return false, nil
			}

			val = *defaultValue
		}
	}

	if reflect.TypeOf(fieldPointer).Elem().Kind() == reflect.Ptr {
		value := reflect.New(reflect.TypeOf(fieldPointer).Elem().Elem())
		reflect.ValueOf(fieldPointer).Elem().Set(value)
		fieldPointer = value.Interface()
	}

	switch typedPointer := fieldPointer.(type) {
	case *string:
		*typedPointer = val
	default:
		_, err := fmt.Sscan(val, fieldPointer)
		if err != nil {
			return false, fmt.Errorf("unable to convert value for environment variable '%s' to target type %t", variableName, reflect.TypeOf(fieldPointer))
		}
	}

	return true, nil
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
