package valid

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/utils"
	"reflect"
	"sync"
)

const (
	SceneAll    Scene = 0
	SceneBind   Scene = 1
	SceneSave   Scene = 10
	SceneInsert Scene = 11
	SceneUpdate Scene = 12
	SceneQuery  Scene = 13
	SceneShow   Scene = 20
	SceneCustom Scene = 1000

	TagRequired Tag = "required"
	TagFormat   Tag = "format"
)

var (
	validate *validator.Validate
	once     sync.Once
	regTypes = &sync.Map{} // 验证注册类型
	regLocs  = &sync.Map{} // 本地化文本缓存
)

type (
	Scene     uint16
	Tag       string
	FieldName string

	// FieldValidRules 定义字段验证规则
	FieldValidRules  = map[Scene]FieldValidRule
	FieldValidRule   = map[Tag]FieldValidRuleFn
	FieldValidRuleFn = func(value reflect.Value, param string) bool

	// ExtraValidRules 定义额外字段验证规则
	ExtraValidRules    = map[Scene]ExtraValidRule
	ExtraValidRule     = map[Tag]ExtraValidRuleInfo
	ExtraValidRuleInfo struct {
		Field   string
		Param   string
		ValidFn func(value interface{}) bool
	}

	// FuncReportError validator.StructLevel.FuncReportError
	FuncReportError = func(field interface{}, fieldName FieldName, tag Tag, param string)

	// LocalizeValidRules 定义本地化的规则映射
	LocalizeValidRules = map[Scene]LocalizeValidRule
	LocalizeValidRule  struct {
		Rule1 map[Tag]map[FieldName]LocalizeValidRuleParam
		Rule2 map[Tag]LocalizeValidRuleParam
	}
	LocalizeValidRuleParam = [3]interface{} // {msg, param, []any}

	// IFieldValidator 定义字段验证接口
	IFieldValidator interface {
		ValidFieldRules() FieldValidRules
	}

	// IExtraValidator 定义额外字段验证接口
	IExtraValidator interface {
		ValidExtraRules(obj any) (utils.KSMap, ExtraValidRules)
	}

	// IStructValidator 定义结构验证接口
	IStructValidator interface {
		ValidStructRules(scene Scene, obj any, fn FuncReportError)
	}

	// ILocalizeValidator 定义本地化错误规则接口
	ILocalizeValidator interface {
		ValidLocalizeRules() LocalizeValidRules
	}

	Validator struct{}
)

func Get() *validator.Validate {
	once.Do(func() {
		opts := []validator.Option{
			validator.WithRequiredStructEnabled(),
		}
		validate = validator.New(opts...)
		// 默认json标签处理
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("json")
			if name == "-" {
				return fld.Name
			}
			return name
		})
	})
	return validate
}

// Valid 根据场景执行验证，并返回本地化错误信息
func (v *Validator) Valid(scene Scene, obj any) *err.CodeErrs {
	typ := reflect.TypeOf(obj)
	// -- 注册验证 --
	if _, ok := regTypes.Load(typ); !ok {
		if e := v.registerValidations(scene, typ, obj); e != nil {
			return e
		}
	}

	// -- 执行验证 --
	if e := Get().Struct(obj); e != nil {
		return v.handleValidationError(e, typ, scene, obj)
	}
	return nil
}

func (v *Validator) registerValidations(scene Scene, typ reflect.Type, obj any) *err.CodeErrs {
	// 处理组合类型的验证规则
	if e := v.registerEmbeddedValidations(scene, typ, obj); e != nil {
		return e
	}

	// -- 字段验证注册 --
	if fv, okk := obj.(IFieldValidator); okk {
		if errs := v.validFields(scene, fv); errs != nil {
			return errs
		}
	}
	_, hasExtra := obj.(IExtraValidator)
	_, hasStruct := obj.(IStructValidator)
	if hasExtra || hasStruct {
		Get().RegisterStructValidation(func(sl validator.StructLevel) {
			cObj := sl.Current().Addr().Interface()
			if hasExtra {
				// -- 额外验证注册 --
				if ev, okk := cObj.(IExtraValidator); okk {
					v.validExtra(cObj, scene, ev, sl)
				}
			}
			if hasStruct {
				// -- 结构验证注册 --
				if sv, okk := cObj.(IStructValidator); okk {
					v.validStruct(cObj, scene, sv, sl)
				}
			}
		}, obj)
	}
	regTypes.Store(typ, true)
	return nil
}

// registerEmbeddedValidations 递归注册组合类型的验证规则
func (v *Validator) registerEmbeddedValidations(scene Scene, typ reflect.Type, obj any) *err.CodeErrs {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// 检查是否是组合类型的字段
		if field.Anonymous {
			fieldType := field.Type
			var embedObj any

			// 处理指针类型的组合字段
			if fieldType.Kind() == reflect.Ptr {
				if !fieldVal.IsNil() {
					embedObj = fieldVal.Interface()
					fieldType = fieldType.Elem()
				}
			} else {
				// 处理非指针类型的组合字段
				embedObj = fieldVal.Addr().Interface()
			}

			// 只处理结构体类型的组合字段
			if fieldType.Kind() == reflect.Struct && embedObj != nil {
				if e := v.registerValidations(scene, reflect.TypeOf(embedObj), embedObj); e != nil {
					return e
				}
			}
		}
	}
	return nil
}

func (v *Validator) validFields(scene Scene, fv IFieldValidator) *err.CodeErrs {
	sceneRules := fv.ValidFieldRules()
	tagRules := make(map[Tag]FieldValidRuleFn)
	if tRules := sceneRules[SceneAll]; tRules != nil {
		for tag, rule := range tRules {
			tagRules[tag] = rule
		}
	}
	if tRules := sceneRules[scene]; tRules != nil {
		for tag, rule := range tRules {
			tagRules[tag] = rule
		}
	}
	for tag, rule := range tagRules {
		if e := Get().RegisterValidation(string(tag), func(fl validator.FieldLevel) bool {
			return rule(fl.Field(), fl.Param())
		}); e != nil {
			return err.Match(e)
		}
	}
	return nil
}

func (v *Validator) validExtra(obj any, scene Scene, ev IExtraValidator, sl validator.StructLevel) {
	extra, sceneRules := ev.ValidExtraRules(obj)
	tagRules := make(map[Tag]ExtraValidRuleInfo)
	if tRules := sceneRules[SceneAll]; tRules != nil {
		for tag, rule := range tRules {
			tagRules[tag] = rule
		}
	}
	if tRules := sceneRules[scene]; tRules != nil {
		for tag, rule := range tRules {
			tagRules[tag] = rule
		}
	}
	for tag, rule := range tagRules {
		value, exists := extra[string(tag)]
		if (tag == TagRequired) && !exists {
			sl.ReportError(value, rule.Field, rule.Field, string(tag), rule.Param)
			continue
		}
		if exists && !rule.ValidFn(value) {
			sl.ReportError(value, rule.Field, rule.Field, string(tag), rule.Param)
		}
	}
}

func (v *Validator) validStruct(obj any, scene Scene, sv IStructValidator, sl validator.StructLevel) {
	sv.ValidStructRules(scene, obj, func(field interface{}, fieldName FieldName, tag Tag, param string) {
		sl.ReportError(field, string(fieldName), string(fieldName), string(tag), param)
	})
}

func (v *Validator) handleValidationError(e error, typ reflect.Type, scene Scene, obj any) *err.CodeErrs {
	var invalidErr *validator.InvalidValidationError
	if errors.As(e, &invalidErr) {
		return err.Match(e)
	}
	var validateErrs validator.ValidationErrors
	if errors.As(e, &validateErrs) {
		// -- 本地化错误注册 --
		if rl, ok := obj.(ILocalizeValidator); ok {
			return v.validLocalize(typ, scene, rl, validateErrs)
		}
	}
	return err.Match(e)
}

func (v *Validator) validLocalize(typ reflect.Type, scene Scene, rl ILocalizeValidator, validateErrs validator.ValidationErrors) *err.CodeErrs {
	tagFieldRules := make(map[Tag]map[FieldName]LocalizeValidRuleParam)
	tagRules := make(map[Tag]LocalizeValidRuleParam)
	cacheRules, ok := regLocs.Load(typ)
	if !ok {
		sceneRules := rl.ValidLocalizeRules()
		if tRules := sceneRules[SceneAll]; tRules.Rule1 != nil {
			for tag, rule := range tRules.Rule1 {
				tagFieldRules[tag] = rule
			}
		}
		if tRules := sceneRules[SceneAll]; tRules.Rule2 != nil {
			for tag, rule := range tRules.Rule2 {
				tagRules[tag] = rule
			}
		}
		if tRules := sceneRules[scene]; tRules.Rule1 != nil {
			for tag, rule := range tRules.Rule1 {
				tagFieldRules[tag] = rule
			}
		}
		if tRules := sceneRules[scene]; tRules.Rule2 != nil {
			for tag, rule := range tRules.Rule2 {
				tagRules[tag] = rule
			}
		}
		regLocs.Store(typ, LocalizeValidRule{tagFieldRules, tagRules})
	} else {
		tagFieldRules = cacheRules.(LocalizeValidRule).Rule1
		tagRules = cacheRules.(LocalizeValidRule).Rule2
	}

	cErrs := err.New()
	for _, ee := range validateErrs {
		// -- 本地化错误注册(Tag+Field) --
		for tag, fieldRules := range tagFieldRules {
			if ee.Tag() == string(tag) {
				for field, rules := range fieldRules {
					if ee.Field() == string(field) {
						var params []any
						if rules[2] != nil {
							params = append(params, rules[2].([]any)...)
						}
						if rules[1].(bool) {
							params = append(params, ee.Param())
						}
						_ = cErrs.WrapLocalize(rules[0].(string), params, nil)
					}
				}
			}
		}
		// -- 本地化错误注册(Tag) --
		for tag, rules := range tagRules {
			if ee.Tag() == string(tag) {
				var params []any
				if rules[2] != nil {
					params = append(params, rules[2].([]any)...)
				}
				if rules[1].(bool) {
					params = append(params, ee.Param())
				}
				_ = cErrs.WrapLocalize(rules[0].(string), params, nil)
			}
		}
	}
	if cErrs.IsNil() {
		_ = cErrs.WrapLocalize("unknown_err", nil, nil)
	}
	return cErrs.Real()
}
