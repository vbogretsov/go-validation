package rule

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/vbogretsov/go-validation"
)

func unexpectedType(v interface{}) validation.Panic {
	return validation.Panic{
		Err: fmt.Sprintf("unexpected type: %v", reflect.TypeOf(v)),
	}
}

// NotNil creates validator to check whether a value is nil.
func NotNil(msg string) validation.Rule {
	return func(v interface{}) error {
		t := reflect.TypeOf(v)
		if t.Kind() != reflect.Ptr {
			return unexpectedType(v)
		}

		switch t.Elem().Kind() {
		case reflect.Interface,
			reflect.Ptr,
			reflect.Slice,
			reflect.Func,
			reflect.Map,
			reflect.Chan:

			if reflect.ValueOf(v).Elem().IsNil() {
				return errors.New(msg)
			}
		default:
			return unexpectedType(v)
		}

		return nil
	}
}

// In creates a validator to chech wheter an item belongs to the set provided.
func In(values []interface{}, msg string) validation.Rule {
	set := map[interface{}]bool{}
	for _, v := range values {
		set[v] = true
	}

	return func(v interface{}) error {
		if !set[v] {
			return fmt.Errorf(msg, v, values)
		}
		return nil
	}
}
