package valid

import (
	"errors"
	"github.com/go-playground/validator/v10"
	"reflect"
	"sync"
)

const (
	SceneAll    Scene = 0
	SceneBind   Scene = 1
	SceneSave   Scene = 10 // = insert + update
	SceneInsert Scene = 11
	SceneUpdate Scene = 12
	SceneQuery  Scene = 13
	SceneReturn Scene = 20 // = response
	SceneCustom Scene = 1000

	TagRequired Tag = "required"
	TagFormat   Tag = "format"
	TagRange    Tag = "range"
	TagCheck    Tag = "check"
)

var (
	v    *Validator
	once sync.Once
)

type (
	Validator struct {
		validate *validator.Validate
		regTypes *sync.Map // 验证注册类型
		regLocs  *sync.Map // 本地化文本缓存
	}

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
		ValidFn func(value any) bool
	}

	// FuncReportError validator.StructLevel.FuncReportError
	FuncReportError = func(field any, fieldName FieldName, tag Tag, param string)

	// LocalizeValidRules 定义本地化的规则映射
	LocalizeValidRules = map[Scene]LocalizeValidRule
	LocalizeValidRule  struct {
		Rule1 map[Tag]map[FieldName]LocalizeValidRuleParam
		Rule2 map[Tag]LocalizeValidRuleParam
	}
	LocalizeValidRuleParam = [3]any // {msg, param, template([]any)}

	// MsgErr 定义错误信息结构体
	MsgErr struct {
		Err    error
		Msg    string
		Params []any
	}

	// IFieldValidator 定义字段验证接口
	IFieldValidator interface {
		ValidFieldRules() FieldValidRules
	}

	// IExtraValidator 定义额外字段验证接口
	IExtraValidator interface {
		ValidExtraRules() (map[string]any, ExtraValidRules)
	}

	// IStructValidator 定义结构验证接口
	IStructValidator interface {
		ValidStructRules(scene Scene, fn FuncReportError)
	}

	// ILocalizeValidator 定义本地化错误规则接口
	ILocalizeValidator interface {
		ValidLocalizeRules() LocalizeValidRules
	}
)

func Get() *Validator {
	once.Do(func() {
		opts := []validator.Option{
			validator.WithRequiredStructEnabled(),
		}
		v = &Validator{
			validate: validator.New(opts...),
			regTypes: &sync.Map{},
			regLocs:  &sync.Map{},
		}
		// 设置Tag <- 默认json标签
		//validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		//	name := fld.Tag.Get("json")
		//	if name == "-" {
		//		return fld.Name
		//	}
		//	return name
		//})
	})
	return v
}

// Check 根据场景执行验证，并返回本地化错误信息
func Check(obj any, scene Scene) []*MsgErr {
	v := Get()
	typ := reflect.TypeOf(obj)

	// -- 注册验证 --
	if _, ok := v.regTypes.Load(typ); !ok {
		if e := v.registerValidations(obj, scene); e != nil {
			return []*MsgErr{{Err: e}}
		}
		v.regTypes.Store(typ, true)
	}

	// -- 执行验证 --
	if e := v.validate.Struct(obj); e != nil {
		return v.handleValidationError(obj, typ, scene, e)
	}
	return nil
}

func (v *Validator) registerValidations(obj any, scene Scene) error {
	// -- 字段验证注册 --
	if e := v.validFields(obj, scene); e != nil {
		return e
	}
	v.validate.RegisterStructValidation(func(sl validator.StructLevel) {
		cObj := sl.Current().Addr().Interface()
		// -- 额外验证注册 --
		v.validExtra(cObj, sl, scene)
		// -- 结构验证注册 --
		v.validStruct(cObj, sl, scene)
	}, obj)
	return nil
}

func (v *Validator) validFields(obj any, scene Scene) error {
	// 处理嵌入字段的验证规则
	if e := v.registerEmbeddedValidations(obj, reflect.TypeOf(obj), scene, 1, nil); e != nil {
		return e
	}

	fv, ok := obj.(IFieldValidator)
	if !ok {
		return nil
	}

	sceneRules := fv.ValidFieldRules()
	if sceneRules == nil {
		return nil
	}
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
		if e := v.validate.RegisterValidation(string(tag), func(fl validator.FieldLevel) bool {
			return rule(fl.Field(), fl.Param())
		}); e != nil {
			return e
		}
	}
	return nil
}

func (v *Validator) validExtra(obj any, sl validator.StructLevel, scene Scene) {
	// 处理嵌入字段的验证规则
	_ = v.registerEmbeddedValidations(obj, reflect.TypeOf(obj), scene, 2, sl)

	ev, ok := obj.(IExtraValidator)
	if !ok {
		return
	}

	extra, sceneRules := ev.ValidExtraRules()
	if (extra == nil) || (sceneRules == nil) {
		return
	}
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

func (v *Validator) validStruct(obj any, sl validator.StructLevel, scene Scene) {
	// 处理嵌入字段的验证规则
	_ = v.registerEmbeddedValidations(obj, reflect.TypeOf(obj), SceneAll, 3, sl)
	_ = v.registerEmbeddedValidations(obj, reflect.TypeOf(obj), scene, 3, sl)

	sv, ok := obj.(IStructValidator)
	if !ok {
		return
	}

	sv.ValidStructRules(SceneAll, func(field any, fieldName FieldName, tag Tag, param string) {
		sl.ReportError(field, string(fieldName), string(fieldName), string(tag), param)
	})
	sv.ValidStructRules(scene, func(field any, fieldName FieldName, tag Tag, param string) {
		sl.ReportError(field, string(fieldName), string(fieldName), string(tag), param)
	})
}

// registerEmbeddedValidations 递归注册组合类型的验证规则
func (v *Validator) registerEmbeddedValidations(
	obj any,
	typ reflect.Type,
	scene Scene,
	ttt int,
	sl validator.StructLevel,
) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		// 检查是否是组合类型的字段
		field := typ.Field(i)
		if !field.Anonymous {
			continue
		}

		fieldVal := val.Field(i)
		fieldType := field.Type
		var embedObj any

		// 处理指针类型的组合字段
		if fieldType.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				continue
			}
			embedObj = fieldVal.Interface()
			fieldType = fieldType.Elem()
		} else {
			// 处理非指针类型的组合字段
			embedObj = fieldVal.Addr().Interface()
		}

		// 只处理结构体类型的组合字段
		if fieldType.Kind() != reflect.Struct || embedObj == nil {
			continue
		}
		switch ttt {
		case 1:
			if fv, okk := embedObj.(IFieldValidator); okk {
				return v.validFields(fv, scene)
			}
		case 2:
			if _, okk := embedObj.(IExtraValidator); okk {
				v.validExtra(embedObj, sl, scene)
			}
		case 3:
			if _, okk := embedObj.(IStructValidator); okk {
				v.validStruct(embedObj, sl, scene)
			}
		}
	}
	return nil
}

func (v *Validator) handleValidationError(
	obj any,
	typ reflect.Type,
	scene Scene,
	e error,
) []*MsgErr {
	var invalidErr *validator.InvalidValidationError
	if errors.As(e, &invalidErr) {
		// -- 验证失败 --
		return []*MsgErr{{Err: e, Msg: "invalid_object_validation"}}
	}
	var validateErrs validator.ValidationErrors
	if errors.As(e, &validateErrs) {
		// -- 本地化错误注册 --
		if rl, ok := obj.(ILocalizeValidator); ok {
			return v.validLocalize(scene, typ, obj, rl, validateErrs, true)
		}
	}
	return []*MsgErr{{Err: e, Msg: "unknown_validator_err"}}
}

func (v *Validator) validLocalize(
	scene Scene,
	typ reflect.Type,
	obj any,
	rl ILocalizeValidator,
	validateErrs validator.ValidationErrors,
	first bool,
) []*MsgErr {
	var msgErrs []*MsgErr
	// 处理组合类型的验证规则
	if msgEs := v.registerEmbeddedLocalizes(scene, typ, obj, validateErrs); msgEs != nil {
		msgErrs = append(msgErrs, msgEs...)
	}

	tagFieldRules := make(map[Tag]map[FieldName]LocalizeValidRuleParam)
	tagRules := make(map[Tag]LocalizeValidRuleParam)
	cacheRules, ok := v.regLocs.Load(typ)
	if !ok {
		sceneRules := rl.ValidLocalizeRules()
		if sceneRules == nil {
			return msgErrs
		}
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
		v.regLocs.Store(typ, LocalizeValidRule{tagFieldRules, tagRules})
	} else {
		tagFieldRules = cacheRules.(LocalizeValidRule).Rule1
		tagRules = cacheRules.(LocalizeValidRule).Rule2
	}

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
						msgErrs = append(msgErrs, &MsgErr{Msg: rules[0].(string), Params: params})
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
				msgErrs = append(msgErrs, &MsgErr{Msg: rules[0].(string), Params: params})
			}
		}
	}
	if (len(msgErrs) <= 0) && first {
		msgErrs = append(msgErrs, &MsgErr{Msg: "unknown_validator_err"})
	}
	return msgErrs
}

// registerEmbeddedLocalizes 递归注册组合类型的本地化规则
func (v *Validator) registerEmbeddedLocalizes(
	scene Scene,
	typ reflect.Type,
	obj any,
	validateErrs validator.ValidationErrors,
) []*MsgErr {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	// 遍历所有字段
	for i := 0; i < typ.NumField(); i++ {
		// 检查是否是组合类型的字段
		field := typ.Field(i)
		if !field.Anonymous {
			continue
		}

		fieldVal := val.Field(i)
		fieldType := field.Type
		var embedObj any

		// 处理指针类型的组合字段
		if fieldType.Kind() == reflect.Ptr {
			if fieldVal.IsNil() {
				continue
			}
			embedObj = fieldVal.Interface()
			fieldType = fieldType.Elem()
		} else {
			// 处理非指针类型的组合字段
			embedObj = fieldVal.Addr().Interface()
		}

		// 只处理实现了 ILocalizeValidator 接口的组合字段
		if fieldType.Kind() != reflect.Struct || embedObj == nil {
			continue
		}
		if embedLocValidator, ok := embedObj.(ILocalizeValidator); ok {
			if msgErrs := v.validLocalize(
				scene,
				reflect.TypeOf(embedObj),
				embedObj,
				embedLocValidator,
				validateErrs,
				false,
			); msgErrs != nil {
				return msgErrs
			}
		}
	}
	return nil
}
