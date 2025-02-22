package valid

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	tagSplit   = ","
	fieldTag   = "valid"
	requireTag = "required"
	groupTag   = "group:"

	cache           sync.Map
	fieldValidators = map[string]ValidatorFieldFunc{}
	groupValidators = map[string]ValidatorGroupFunc{}
)

type ValidatorFieldFunc func(value interface{}) (string, bool)

type ValidatorGroupFunc func(map[string]interface{}) (string, bool)

// FieldValidator 使用字段索引代替字段名，优化字段查找
type FieldValidator struct {
	Index      []int // 使用字段索引数组替代字段名
	Name       string
	Required   bool
	Validators []ValidatorFieldFunc
	ZeroCheck  func(reflect.Value) bool // 编译期确定的零值检查函数
}

// GroupValidator 专门处理组验证
type GroupValidator struct {
	Fields    []string                    // 组内字段名
	Validator func(reflect.Value) []error // 组验证函数
}

type StructValidator struct {
	FieldValidators []FieldValidator
	GroupValidators map[string]GroupValidator // 组验证器映射
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

	tv := &StructValidator{
		FieldValidators: make([]FieldValidator, 0),
		GroupValidators: make(map[string]GroupValidator),
	}

	groupFields := make(map[string][]string)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(fieldTag)
		if tag == "" {
			continue
		}

		rules := strings.Split(tag, tagSplit)
		var fValidators []ValidatorFieldFunc
		required := false

		for _, rule := range rules {
			rule = strings.TrimSpace(rule)
			if rule == requireTag {
				required = true
			} else if strings.HasPrefix(rule, groupTag) {
				groupName := strings.TrimPrefix(rule, groupTag)
				groupFields[groupName] = append(groupFields[groupName], field.Name)
			} else if rule != "" {
				if validator, exists := fieldValidators[rule]; exists {
					fValidators = append(fValidators, validator)
				}
			}
		}

		if required || len(fValidators) > 0 {
			fv := FieldValidator{
				Index:      field.Index,
				Name:       field.Name,
				Required:   required,
				Validators: fValidators,
				ZeroCheck:  getZeroChecker(field.Type.Kind()),
			}
			tv.FieldValidators = append(tv.FieldValidators, fv)
		}
	}

	for groupName, fields := range groupFields {
		tv.GroupValidators[groupName] = GroupValidator{
			Fields: fields,
			Validator: func(v reflect.Value) []error {
				var errs []error
				// 如果有自定义的组验证器，就使用它
				if groupValidator, exists := groupValidators[groupName]; exists {
					fieldValues := make(map[string]interface{})
					for _, fieldName := range fields {
						field := v.FieldByName(fieldName)
						fieldValues[fieldName] = field.Interface()
					}
					if msg, ok := groupValidator(fieldValues); !ok {
						errs = append(errs, fmt.Errorf(msg))
					}
					return errs
				}

				// 默认的组验证逻辑（全空或全填）
				allEmpty := true
				allFilled := true

				for _, fieldName := range fields {
					field := v.FieldByName(fieldName)
					isEmpty := getZeroChecker(field.Kind())(field)
					if isEmpty {
						allFilled = false
					} else {
						allEmpty = false
					}
				}

				if !allEmpty && !allFilled {
					errs = append(errs, fmt.Errorf("group %s: all fields must be either all empty or all filled", groupName))
				}
				return errs
			},
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
	if tv == nil || (len(tv.FieldValidators) == 0 && len(tv.GroupValidators) == 0) {
		return nil
	}

	// 验证单个字段
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

	// 执行组验证
	for _, gv := range tv.GroupValidators {
		if es := gv.Validator(val); len(es) > 0 {
			errs = append(errs, es...)
		}
	}
	return errs
}

func RegisterStructs(
	fieldValids map[string]ValidatorFieldFunc,
	groupValids map[string]ValidatorGroupFunc,
	vs []interface{},
) {
	fieldValidators = fieldValids
	groupValidators = groupValids // 添加组验证器注册
	for _, v := range vs {
		t := reflect.TypeOf(v)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		CompileValidators(t)
	}
}
