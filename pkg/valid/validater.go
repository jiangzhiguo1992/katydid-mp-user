package valid

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

var (
	cache      sync.Map
	fieldTag   = "valid"
	requireTag = "required"
	validators = map[string]ValidatorFunc{}
)

type ValidatorFunc func(value interface{}) (string, bool)

// FieldValidator 使用字段索引代替字段名，优化字段查找
type FieldValidator struct {
	Index      []int // 使用字段索引数组替代字段名
	Name       string
	Required   bool
	Validators []ValidatorFunc
	ZeroCheck  func(reflect.Value) bool // 编译期确定的零值检查函数
}

type StructValidator struct {
	FieldValidators []FieldValidator
}

// 为常见类型生成专用的零值检查函数
func getZeroChecker(kind reflect.Kind) func(reflect.Value) bool {
	switch kind {
	case reflect.String:
		return func(v reflect.Value) bool { return v.Len() == 0 }
	case reflect.Bool:
		return func(v reflect.Value) bool { return !v.Bool() }
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value) bool { return v.Int() == 0 }
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(v reflect.Value) bool { return v.Uint() == 0 }
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value) bool { return v.Float() == 0 }
	case reflect.Slice, reflect.Map:
		return func(v reflect.Value) bool { return v.Len() == 0 }
	case reflect.Interface, reflect.Ptr:
		return func(v reflect.Value) bool { return v.IsNil() }
	default:
		return func(v reflect.Value) bool { return false }
	}
}

func CompileValidators(t reflect.Type) *StructValidator {
	if v, ok := cache.Load(t); ok {
		return v.(*StructValidator)
	}

	tv := &StructValidator{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(fieldTag)
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, ",")
		var fieldValidators []ValidatorFunc
		required := false

		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == requireTag {
				required = true
			} else if rule != "" {
				if validator, exists := validators[rule]; exists {
					fieldValidators = append(fieldValidators, validator)
				}
			}
		}

		if required || len(fieldValidators) > 0 {
			tv.FieldValidators = append(tv.FieldValidators, FieldValidator{
				Index:      field.Index,
				Name:       field.Name,
				Required:   required,
				Validators: fieldValidators,
				ZeroCheck:  getZeroChecker(field.Type.Kind()),
			})
		}
	}

	cache.Store(t, tv)
	return tv
}

func ValidateStruct(v interface{}) []error {
	var errs []error
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	tv := CompileValidators(val.Type())
	if tv == nil || len(tv.FieldValidators) == 0 {
		return nil
	}

	for _, fv := range tv.FieldValidators {
		// 使用字段索引直接获取字段值
		field := val.FieldByIndex(fv.Index)
		if !field.IsValid() {
			continue
		}

		// 使用编译期生成的零值检查函数
		if fv.Required && fv.ZeroCheck(field) {
			errs = append(errs, errors.New("field "+fv.Name+" is required"))
			continue
		}

		fieldValue := field.Interface()
		for _, validator := range fv.Validators {
			if msg, ok := validator(fieldValue); !ok {
				errs = append(errs, errors.New(msg))
			}
		}
	}
	return errs
}

func RegisterStructs(tag string, valids map[string]ValidatorFunc, vs []interface{}) {
	if len(tag) > 0 {
		fieldTag = tag
	}
	validators = valids
	for _, v := range vs {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		CompileValidators(t)
	}
}
