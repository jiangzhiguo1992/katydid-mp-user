package valid

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

var (
	cache      sync.Map                     // 验证器缓存
	fieldTag   = "valid"                    // 自定义标签
	validators = map[string]ValidatorFunc{} // 预定义的验证器
)

// ValidatorFunc 定义验证函数类型，返回错误消息与验证是否通过的标志
type ValidatorFunc func(value interface{}) (string, bool)

// FieldValidator 表示单个字段的验证器，包括必需标识和其他验证函数
type FieldValidator struct {
	Field      string
	Required   bool
	Validators []ValidatorFunc
}

// StructValidator 缓存结构体验证规则
type StructValidator struct {
	FieldValidators []FieldValidator
}

// CompileValidators 预编译类型的验证器，支持必需字段的验证
func CompileValidators(t reflect.Type) *StructValidator {
	// 检查缓存
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
			// 处理必需规则
			if rule == "required" {
				required = true
			} else if rule != "" {
				if validator, exists := validators[rule]; exists {
					fieldValidators = append(fieldValidators, validator)
				}
			}
		}

		// 只要设置了必需或其它验证器，则添加到预编译验证器中
		if required || len(fieldValidators) > 0 {
			tv.FieldValidators = append(tv.FieldValidators, FieldValidator{
				Field:      field.Name,
				Required:   required,
				Validators: fieldValidators,
			})
		}
	}

	cache.Store(t, tv)
	return tv
}

// isZero 判断字段值是否为零值
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// ValidateStruct 使用预编译的验证器验证结构体，增加必需字段检查
func ValidateStruct(v interface{}) []error {
	var errs []error
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}

	// 获取预编译的验证器
	tv := CompileValidators(val.Type())
	if tv == nil || len(tv.FieldValidators) == 0 {
		return nil
	}

	// 反射获取每个字段并执行验证
	for _, fv := range tv.FieldValidators {
		field := val.FieldByName(fv.Field)
		if !field.IsValid() {
			continue
		}

		// 必需字段判断，如果字段值为零则返回错误
		if fv.Required && isZero(field) {
			errs = append(errs, errors.New("field "+fv.Field+" is required"))
			// 如果未通过必需验证，可跳过后续验证
			continue
		}

		// 执行其他验证器
		fieldValue := field.Interface()
		for _, validator := range fv.Validators {
			if msg, ok := validator(fieldValue); !ok {
				errs = append(errs, errors.New(msg))
			}
		}
	}
	return errs
}

// RegisterStructs 注册需要验证的结构体，提前编译验证器
func RegisterStructs(tag string, vs []interface{}) {
	if len(tag) > 0 {
		fieldTag = tag
	}
	for _, v := range vs {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		CompileValidators(t)
	}
}
