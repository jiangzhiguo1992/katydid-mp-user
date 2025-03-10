package valid

//import (
//	"errors"
//	"fmt"
//	"reflect"
//	"sort"
//	"strings"
//	"sync"
//	"sync/atomic"
//)
//
//var (
//	tagSplit   = ","        // 标签分隔符
//	fieldTag   = "valid"    // 字段验证标签
//	requireTag = "required" // 必填标签
//	groupTag   = "group:"   // 分组标签
//
//	cache        sync.Map      // 缓存
//	cacheSize    int32    = 0  // 缓存大小
//	maxCacheSize int32    = -1 // -1表示不限制
//
//	requireTips     = map[string]string{}             // 空提示
//	fieldValidators = map[string]ValidatorFieldFunc{} // 字段验证器注册
//	groupValidators = map[string]ValidatorGroupFunc{} // 分组验证器注册
//)
//
//type ValidatorFieldFunc func(value any) (string, bool)
//type ValidatorGroupFunc func(map[string]any) (string, bool)
//
//type ValidateFunc func(reflect.Value) []error
//type TypeChecker func(any) reflect.Value
//
//// FieldValidator 使用字段索引代替字段名，优化字段查找
//type FieldValidator struct {
//	Index      []int // 使用字段索引数组替代字段名
//	Name       string
//	Required   bool
//	NilTip     string
//	Validators []ValidatorFieldFunc
//	ZeroCheck  func(reflect.Value) bool // 编译期确定的零值检查函数
//}
//
//// GroupValidator 专门处理组验证
//type GroupValidator struct {
//	Indices   [][]int                       // 存储字段索引
//	Fields    []string                      // 组内字段名
//	Validator func([]reflect.Value) []error // 直接使用 reflect.Value
//}
//
//type StructValidator struct {
//	FieldValidators []FieldValidator
//	GroupValidators map[string]GroupValidator
//	validateFunc    ValidateFunc
//	typeChecker     TypeChecker
//}
//
//// 为常见类型生成专用的零值检查函数
//func getZeroChecker(kind reflect.Kind) func(reflect.Value) bool {
//	switch kind {
//	case reflect.String:
//		return func(v reflect.Value) bool { return v.Len() == 0 }
//	//case reflect.Bool:
//	//	return func(v reflect.Value) bool { return !v.Bool() }
//	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//		return func(v reflect.Value) bool { return v.Int() == 0 }
//	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
//		return func(v reflect.Value) bool { return v.Uint() == 0 }
//	case reflect.Float32, reflect.Float64:
//		return func(v reflect.Value) bool { return v.Float() == 0 }
//	case reflect.Slice, reflect.Map:
//		return func(v reflect.Value) bool { return v.Len() == 0 }
//	case reflect.Interface, reflect.Ptr:
//		return func(v reflect.Value) bool { return v.IsNil() }
//	default:
//		return func(v reflect.Value) bool { return false }
//	}
//}
//
//func compileTypeChecker(t reflect.Type) TypeChecker {
//	return func(v any) reflect.Value {
//		val := reflect.ValueOf(v)
//		if val.Kind() == reflect.Pointer {
//			val = val.Elem()
//		}
//		if val.Type() != t {
//			panic(fmt.Sprintf("■ ■ valid ■ ■ invalid type: expected %v, got %v", t, val.Type()))
//		}
//		return val
//	}
//}
//
//func compileValidateFunc(sv *StructValidator) ValidateFunc {
//	return func(val reflect.Value) []error {
//		var errs []error
//
//		// 验证单个字段
//		for _, fv := range sv.FieldValidators {
//			field := val.FieldByIndex(fv.Index)
//			if !field.IsValid() {
//				continue
//			}
//
//			// 使用编译期生成的零值检查函数
//			if fv.Required && fv.ZeroCheck(field) {
//				errs = append(errs, errors.New(fv.NilTip))
//				continue
//			}
//
//			fieldValue := field.Interface()
//			for _, validator := range fv.Validators {
//				if msg, ok := validator(fieldValue); !ok {
//					errs = append(errs, errors.New(msg))
//				}
//			}
//		}
//
//		// 执行组验证
//		for _, gv := range sv.GroupValidators {
//			values := make([]reflect.Value, len(gv.Indices))
//			for i, idx := range gv.Indices {
//				values[i] = val.FieldByIndex(idx)
//			}
//			if es := gv.Validator(values); len(es) > 0 {
//				errs = append(errs, es...)
//			}
//		}
//		return errs
//	}
//}
//
//func addToCache(t reflect.Type, sv *StructValidator) {
//	if maxCacheSize >= 0 {
//		if atomic.LoadInt32(&cacheSize) >= maxCacheSize {
//			return
//		}
//	}
//	cache.Store(t, sv)
//	atomic.AddInt32(&cacheSize, 1)
//}
//
//func compileValidators(t reflect.Type) *StructValidator {
//	if t.Kind() == reflect.Pointer {
//		t = t.Elem()
//	}
//
//	// 检查缓存
//	if v, ok := cache.Load(t); ok {
//		return v.(*StructValidator)
//	}
//
//	tv := &StructValidator{
//		FieldValidators: make([]FieldValidator, 0),
//		GroupValidators: make(map[string]GroupValidator),
//	}
//
//	groupFields := make(map[string][]string)
//	for i := 0; i < t.NumField(); i++ {
//		field := t.Field(i)
//
//		// 处理嵌套结构体
//		fieldType := field.Type
//		if fieldType.Kind() == reflect.Struct {
//			nested := compileValidators(fieldType)
//			if nested != nil {
//				// 调整嵌套字段的索引
//				for _, fv := range nested.FieldValidators {
//					newIndex := append([]int{i}, fv.Index...)
//					fieldName := field.Name + "." + fv.Name
//					nilTip := ""
//					if tip, ok := requireTips[fieldName]; ok {
//						nilTip = tip
//					}
//					tv.FieldValidators = append(tv.FieldValidators, FieldValidator{
//						Index:      newIndex,
//						Name:       fieldName,
//						Required:   fv.Required,
//						NilTip:     nilTip,
//						Validators: fv.Validators,
//						ZeroCheck:  fv.ZeroCheck,
//					})
//				}
//				// 合并组验证器
//				for groupName, gv := range nested.GroupValidators {
//					tv.GroupValidators[field.Name+"."+groupName] = gv
//				}
//			}
//		}
//
//		// 处理当前字段的验证规则
//		tag := field.Tag.Get(fieldTag)
//		if tag == "" {
//			continue
//		}
//
//		rules := strings.Split(tag, tagSplit)
//		var fValidators []ValidatorFieldFunc
//		required := false
//
//		for _, rule := range rules {
//			rule = strings.TrimSpace(rule)
//			if rule == requireTag {
//				required = true
//			} else if strings.HasPrefix(rule, groupTag) {
//				groupName := strings.TrimPrefix(rule, groupTag)
//				groupFields[groupName] = append(groupFields[groupName], field.Name)
//			} else if rule != "" {
//				if validator, exists := fieldValidators[rule]; exists {
//					fValidators = append(fValidators, validator)
//				}
//			}
//		}
//
//		if required || len(fValidators) > 0 {
//			nilTip := ""
//			if required {
//				if tip, ok := requireTips[field.Name]; ok {
//					nilTip = tip
//				}
//			}
//
//			fv := FieldValidator{
//				Index:      field.Index, // []int{i},
//				Name:       field.Name,
//				Required:   required,
//				NilTip:     nilTip,
//				Validators: fValidators,
//				ZeroCheck:  getZeroChecker(field.Type.Kind()),
//			}
//			tv.FieldValidators = append(tv.FieldValidators, fv)
//		}
//	}
//
//	// Require 进行排序
//	sort.Slice(tv.FieldValidators, func(i, j int) bool {
//		return tv.FieldValidators[i].Required && !tv.FieldValidators[j].Required
//	})
//
//	// 处理组验证器
//	for groupName, fields := range groupFields {
//		indices := make([][]int, len(fields))
//		for i, fieldName := range fields {
//			field, _ := t.FieldByName(fieldName)
//			indices[i] = field.Index
//		}
//
//		tv.GroupValidators[groupName] = GroupValidator{
//			Fields:  fields,
//			Indices: indices,
//			Validator: func(values []reflect.Value) []error {
//				var errs []error
//				// 如果有自定义的组验证器，就使用它
//				if groupValidator, exists := groupValidators[groupName]; exists {
//					fieldValues := make(map[string]any, len(fields))
//					for i, fieldName := range fields {
//						fieldValues[fieldName] = values[i].Interface()
//					}
//					if msg, ok := groupValidator(fieldValues); !ok {
//						errs = append(errs, fmt.Errorf(msg))
//					}
//					return errs
//				}
//
//				// 默认的组验证逻辑（全空或全填）
//				/*allEmpty := true
//				allFilled := true
//
//				for _, value := range values {
//					isEmpty := getZeroChecker(value.Kind())(value)
//					if isEmpty {
//						allFilled = false
//					} else {
//						allEmpty = false
//					}
//				}
//
//				if !allEmpty && !allFilled {
//					errs = append(errs, fmt.Errorf("group %s: all fields must be either all empty or all filled", groupName))
//				}*/
//				return errs
//			},
//		}
//	}
//
//	// 在保存到缓存之前，添加预编译的函数
//	tv.typeChecker = compileTypeChecker(t)
//	tv.validateFunc = compileValidateFunc(tv)
//
//	addToCache(t, tv)
//	return tv
//}
//
//func ValidateStruct(v any) []error {
//	val := reflect.ValueOf(v)
//	if val.Kind() == reflect.Pointer {
//		val = val.Elem()
//	}
//
//	tv := compileValidators(val.Type())
//	if tv == nil || (len(tv.FieldValidators) == 0 && len(tv.GroupValidators) == 0) {
//		return nil
//	}
//
//	// 返回预编译的函数的结果
//	return tv.validateFunc(tv.typeChecker(v))
//}
//
//func RegisterRules(
//	nilTips map[string]string,
//	fieldValids map[string]ValidatorFieldFunc,
//	groupValids map[string]ValidatorGroupFunc,
//	structs []any,
//) {
//	fieldTag = "valid"
//	maxCacheSize = -1
//	requireTips = nilTips
//	fieldValidators = fieldValids
//	groupValidators = groupValids
//	for _, v := range structs {
//		t := reflect.TypeOf(v)
//		if t.Kind() == reflect.Pointer {
//			t = t.Elem()
//		}
//		compileValidators(t)
//	}
//}
