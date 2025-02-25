package valid

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/utils"
	"reflect"
	"sync"
)

const (
	SceneAll   uint16 = 0
	SceneBind  uint16 = 1
	SceneSave  uint16 = 2
	SceneQuery uint16 = 3
	SceneShow  uint16 = 4
)

var (
	validate *validator.Validate
	once     sync.Once
	regTypes sync.Map // 避免重复注册
)

type (
	// FuncReportError validator.StructLevel.FuncReportError
	FuncReportError = func(field interface{}, fieldName, structFieldName string, tag, param string)

	// FieldValidResult 定义字段验证规则
	FieldValidResult = map[uint16]map[string]func(reflect.Value) bool

	// ExtraValidResult 定义额外字段验证规则
	ExtraValidResult = map[uint16]map[string]ExtraValidationInfo

	// RulesValidLocalize 定义本地化的规则映射
	RulesValidLocalize = map[uint16]struct {
		Rule1 map[string]map[string][3]interface{}
		Rule2 map[string][3]interface{}
	}

	// ExtraValidationInfo 定义额外验证信息
	ExtraValidationInfo struct {
		Key      string
		Required bool
		Validate func(value interface{}) bool
	}

	// IFieldValidator 定义字段验证接口
	IFieldValidator interface {
		ValidFieldRules() FieldValidResult
	}

	// IExtraValidator 定义额外字段验证接口
	IExtraValidator interface {
		ValidExtraRules() (utils.KSMap, ExtraValidResult)
	}

	// IStructValidator 定义结构验证接口
	IStructValidator interface {
		ValidStructRules(obj any, fn FuncReportError)
	}

	// IRuleLocalizes 定义本地化错误规则接口
	IRuleLocalizes interface {
		ValidRuleLocalizes() RulesValidLocalize
	}

	Validator struct {
		validate *validator.Validate
		any
	}
)

func Get() *validator.Validate {
	once.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("json")
			if name == "-" {
				return fld.Name
			}
			return name
		})
		// TODO:GG trans
	})
	return validate
}

func New(obj any) *Validator {
	return &Validator{
		validate: Get(),
		any:      obj,
	}
}

// Valid 根据场景执行验证，并返回本地化错误信息
func (v *Validator) Valid(scene uint16) *err.CodeErrs {
	typ := reflect.TypeOf(v.any)
	if _, ok := regTypes.Load(typ); !ok {
		// -- 字段验证注册 --
		if fv, ok := v.any.(IFieldValidator); ok {
			scenes := fv.ValidFieldRules()
			sceneRules := make(map[string]func(reflect.Value) bool)
			if allRules := scenes[SceneAll]; allRules != nil {
				for name, rule := range allRules {
					sceneRules[name] = rule
				}
			}
			if specificRules := scenes[scene]; specificRules != nil {
				for name, rule := range specificRules {
					sceneRules[name] = rule
				}
			}
			for name, rule := range sceneRules {
				if e := v.validate.RegisterValidation(name, func(fl validator.FieldLevel) bool {
					return rule(fl.Field())
				}); e != nil {
					return err.Match(e)
				}
			}
		}
		// -- 额外验证注册 --
		if ev, ok := v.any.(IExtraValidator); ok {
			v.validate.RegisterStructValidation(func(sl validator.StructLevel) {
				extra, scenes := ev.ValidExtraRules()
				sceneRules := make(map[string]ExtraValidationInfo)
				if allRules := scenes[SceneAll]; allRules != nil {
					for name, rule := range allRules {
						sceneRules[name] = rule
					}
				}
				if specificRules := scenes[scene]; specificRules != nil {
					for name, rule := range specificRules {
						sceneRules[name] = rule
					}
				}
				for key, rule := range sceneRules {
					value, exists := extra[key]
					if rule.Required && !exists {
						sl.ReportError(extra, "Extra", "Extra", fmt.Sprintf("required-%s", key), "")
						continue
					}
					if exists && !rule.Validate(value) {
						sl.ReportError(extra, "Extra", "Extra", fmt.Sprintf("invalid-%s", key), "")
					}
				}
			}, v)
		}
		// -- 结构验证注册 --
		if sv, ok := v.any.(IStructValidator); ok {
			v.validate.RegisterStructValidation(func(sl validator.StructLevel) {
				sv.ValidStructRules(sl.Current().Interface(), sl.ReportError)
			}, v)
		}
		regTypes.Store(typ, true)
	}

	if e := v.validate.Struct(v.any); e != nil {
		var invalidErr *validator.InvalidValidationError
		if errors.As(e, &invalidErr) {
			return err.Match(e)
		}
		var validateErrs validator.ValidationErrors
		if errors.As(e, &validateErrs) {
			cErrs := err.New()
			if rl, ok := v.any.(IRuleLocalizes); ok {
				scenes := rl.ValidRuleLocalizes()
				sceneRule1s := make(map[string]map[string][3]interface{})
				sceneRule2s := make(map[string][3]interface{})
				if allRules := scenes[SceneAll]; allRules.Rule1 != nil {
					for name, rule := range allRules.Rule1 {
						sceneRule1s[name] = rule
					}
				}
				if allRules := scenes[SceneAll]; allRules.Rule2 != nil {
					for name, rule := range allRules.Rule2 {
						sceneRule2s[name] = rule
					}
				}
				if specificRules := scenes[scene]; specificRules.Rule1 != nil {
					for name, rule := range specificRules.Rule1 {
						sceneRule1s[name] = rule
					}
				}
				if specificRules := scenes[scene]; specificRules.Rule2 != nil {
					for name, rule := range specificRules.Rule2 {
						sceneRule2s[name] = rule
					}
				}
				for _, ee := range validateErrs {
					for kk, vv := range sceneRule1s {
						if ee.Tag() == kk {
							for kkk, vvv := range vv {
								if ee.Field() == kkk {
									var params []any
									if vvv[2] != nil {
										params = append(params, vvv[2].([]any)...)
									}
									if vvv[1].(bool) {
										params = append(params, ee.Param())
									}
									_ = cErrs.WrapLocalize(vvv[0].(string), params, nil)
								}
							}
						}
					}
					for kk, vv := range sceneRule2s {
						if ee.Tag() == kk {
							var params []any
							if vv[2] != nil {
								params = append(params, vv[2].([]any)...)
							}
							if vv[1].(bool) {
								params = append(params, ee.Param())
							}
							_ = cErrs.WrapLocalize(vv[0].(string), params, nil)
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
