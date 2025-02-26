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
	SceneAll   Scene = 0
	SceneBind  Scene = 1
	SceneSave  Scene = 2
	SceneQuery Scene = 3
	SceneShow  Scene = 4

	TagRequired Tag = "required"
	TagFormat   Tag = "format"
)

var (
	validate *validator.Validate
	once     sync.Once
	regTypes sync.Map // 避免重复注册
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
	FuncReportError = func(field interface{}, fieldName FieldName, tag, param string)

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
		ValidExtraRules() (utils.KSMap, ExtraValidRules)
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
		validate = validator.New(validator.WithRequiredStructEnabled())
		//validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		//	name := fld.Tag.Get("json")
		//	if name == "-" {
		//		return fld.Name
		//	}
		//	return name
		//})
	})
	return validate
}

// Valid 根据场景执行验证，并返回本地化错误信息
func (v *Validator) Valid(scene Scene, obj any) *err.CodeErrs {
	typ := reflect.TypeOf(obj)
	if _, ok := regTypes.Load(typ); !ok {
		if fv, okk := obj.(IFieldValidator); okk {
			// -- 字段验证注册 --
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
		}
		ev, okk1 := obj.(IExtraValidator)
		sv, okk2 := obj.(IStructValidator)
		if okk1 || okk2 {
			Get().RegisterStructValidation(func(sl validator.StructLevel) {
				// -- 额外验证注册 --
				if okk1 {
					extra, sceneRules := ev.ValidExtraRules()
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
				// -- 结构验证注册 --
				if okk2 {
					sv.ValidStructRules(scene, sl.Current().Interface(), func(field interface{}, fieldName FieldName, tag, param string) {
						sl.ReportError(field, string(fieldName), string(fieldName), tag, param)
					})
				}
			}, obj)
		}
		regTypes.Store(typ, true)
	}

	// -- 执行验证 --
	if e := Get().Struct(obj); e != nil {
		var invalidErr *validator.InvalidValidationError
		if errors.As(e, &invalidErr) {
			return err.Match(e)
		}
		var validateErrs validator.ValidationErrors
		if errors.As(e, &validateErrs) {
			// TODO:GG 缓存
			cErrs := err.New()
			if rl, ok := obj.(ILocalizeValidator); ok {
				sceneRules := rl.ValidLocalizeRules()
				tagFieldRules := make(map[Tag]map[FieldName]LocalizeValidRuleParam)
				tagRules := make(map[Tag]LocalizeValidRuleParam)
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
			}
			return cErrs
		}
		return err.Match(e)
	}
	return nil
}
