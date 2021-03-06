package rule

import (
	"fmt"
	"reflect"

	"github.com/vbogretsov/go-validation"
)

const (
	eNotPtr = "expected pointer at index %d"
)

var (
	ParamSliceMinLen = "minLen"
	ParamSliceMaxLen = "maxLen"
)

// SliceIter represent iterator interface for slice.
type SliceIter func(interface{}, int) interface{}

func sliceRule(fn validation.Rule) validation.Rule {
	return func(ctx interface{}) func(interface{}) error {
		return func(v interface{}) error {
			t := reflect.TypeOf(v)
			if t.Kind() != reflect.Ptr {
				return unexpectedType(v)
			}

			switch t.Elem().Kind() {
			case reflect.Slice:
				return fn(ctx)(v)
			default:
				return unexpectedType(v)
			}
		}
	}
}

// SliceLen creates validator to check whether slice length is in the range
// provided.
func SliceLen(min, max int, msg string) validation.Rule {
	return sliceRule(wrap(func(v interface{}) error {
		n := reflect.ValueOf(v).Elem().Len()
		if n < min || n > max {
			return validation.Error{
				Message: msg,
				Params: validation.Params{
					ParamSliceMinLen: min,
					ParamSliceMaxLen: max,
				},
			}
		}
		return nil
	}))
}

// SliceMinLen creates validator to check whether slice length is not less than
// the value provided.
func SliceMinLen(min int, msg string) validation.Rule {
	return sliceRule(wrap(func(v interface{}) error {
		n := reflect.ValueOf(v).Elem().Len()
		if n < min {
			return validation.Error{
				Message: msg,
				Params: validation.Params{
					ParamSliceMinLen: min,
				},
			}
		}
		return nil
	}))
}

// SliceMaxLen creates validator to check whether slice length is not less than
// the value provided.
func SliceMaxLen(max int, msg string) validation.Rule {
	return sliceRule(wrap(func(v interface{}) error {
		n := reflect.ValueOf(v).Elem().Len()
		if n > max {
			return validation.Error{
				Message: msg,
				Params: validation.Params{
					ParamSliceMaxLen: max,
				},
			}
		}
		return nil
	}))
}

// SliceEach creates validator to check whether all items of a slice meet the
// rules provided.
func SliceEach(iter SliceIter, rules []validation.Rule) validation.Rule {
	return sliceRule(func(ctx interface{}) func(interface{}) error {
		return func(v interface{}) error {
			ses := []error{}

			n := reflect.ValueOf(v).Elem().Len()
			for i := 0; i < n; i++ {
				se := []error{}
				k := iter(v, i)

				for _, r := range rules {
					if e := r(ctx)(k); e != nil {
						if _, ok := e.(validation.Panic); ok {
							return e
						} else if es, ok := e.(validation.Errors); ok {
							se = append(se, []error(es)...)
						} else {
							se = append(se, e)
						}
					}
				}

				if len(se) > 0 {
					ses = append(ses, validation.SliceError{
						Index:  i,
						Errors: se,
					})
				}
			}

			if len(ses) > 0 {
				return validation.Errors(ses)
			}

			return nil
		}
	})
}

// SliceUnique create validator to check wheter a slice contains only unique
// items.
func SliceUnique(iter SliceIter, msg string) validation.Rule {
	return sliceRule(wrap(func(v interface{}) error {
		errs := []error{}
		set := map[interface{}]bool{}

		n := reflect.ValueOf(v).Elem().Len()
		for i := 0; i < n; i++ {
			p := reflect.ValueOf(iter(v, i))
			if p.Type().Kind() != reflect.Ptr {
				return validation.Panic{Err: fmt.Errorf(eNotPtr, i)}
			}

			k := reflect.Indirect(p).Interface()
			if set[k] {
				errs = append(errs, validation.SliceError{
					Index:  i,
					Errors: []error{validation.Error{Message: msg}},
				})
			}

			set[k] = true
		}

		if len(errs) > 0 {
			return validation.Errors(errs)
		}

		return nil
	}))
}
