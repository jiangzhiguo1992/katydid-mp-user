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

	regTypes = make(map[reflect.Type]bool)
)

type (
	// FuncReportError validator.StructLevel.FuncReportError
	FuncReportError = func(field interface{}, fieldName, structFieldName string, tag, param string)

	FieldValidResult   = map[uint16]map[string]func(reflect.Value) bool
	ExtraValidResult   = map[uint16]map[string]ExtraValidationInfo
	RulesValidLocalize = map[uint16]struct {
		Rule1 map[string]map[string][3]interface{}
		Rule2 map[string][3]interface{}
	}

	ExtraValidationInfo struct {
		Key      string
		Required bool
		Validate func(value interface{}) bool
	}

	IFieldValidator interface {
		ValidFieldRules() FieldValidResult
	}

	IExtraValidator interface {
		ValidExtraRules() (utils.KSMap, ExtraValidResult)
	}

	IStructValidator interface {
		ValidStructRules(obj any, fn FuncReportError)
	}

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

func (v *Validator) Valid(scene uint16) *err.CodeErrs {
	if !regTypes[reflect.TypeOf(v.any)] {
		// fields
		if i, ok := v.any.(IFieldValidator); ok {
			scenes := i.ValidFieldRules()
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
				e := v.validate.RegisterValidation(name, func(fl validator.FieldLevel) bool {
					return rule(fl.Field())
				})
				if e != nil {
					return err.Match(e)
				}
			}
		}

		// extra
		if i, ok := v.any.(IExtraValidator); ok {
			v.validate.RegisterStructValidation(func(sl validator.StructLevel) {
				extra, scenes := i.ValidExtraRules()
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
						sl.ReportError(extra, "Extra", "Extra",
							fmt.Sprintf("required-%s", key), "")
						continue
					}
					if exists && !rule.Validate(value) {
						sl.ReportError(extra, "Extra", "Extra",
							fmt.Sprintf("invalid-%s", key), "")
					}
				}
			}, v)
		}

		// struct
		if i, ok := v.any.(IStructValidator); ok {
			v.validate.RegisterStructValidation(func(sl validator.StructLevel) {
				i.ValidStructRules(sl.Current().Interface(), sl.ReportError)
			}, v)
		}

		// cache
		regTypes[reflect.TypeOf(v.any)] = true
	}

	// localize
	e := v.validate.Struct(v)
	if e != nil {
		var invalidValidationError *validator.InvalidValidationError
		if errors.As(e, &invalidValidationError) {
			return err.Match(e)
		}
		var validateErrs validator.ValidationErrors
		if errors.As(e, &validateErrs) {
			var cErrs = err.New()
			if i, ok := v.any.(IRuleLocalizes); ok {
				scenes := i.ValidRuleLocalizes()
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
	}
	return err.Match(e)
}
