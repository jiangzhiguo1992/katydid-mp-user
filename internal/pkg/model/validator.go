package model

import (
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"katydid-mp-user/pkg/err"
	"katydid-mp-user/pkg/valid"
	"katydid-mp-user/utils"
	"reflect"
)

const (
	ValidSceneAll  uint16 = 0
	ValidSceneBind uint16 = 1
	ValidSceneSave uint16 = 2
)

var (
	regTypes = make(map[reflect.Type]bool)
)

type (
	// ValidReportError validator.StructLevel.ValidReportError
	ValidReportError = func(field interface{}, fieldName, structFieldName string, tag, param string)

	ValidFieldResult  = map[uint16]map[string]func(reflect.Value) bool
	ValidExtraResult  = map[uint16]map[string]ValidationExtraInfo
	ValidRuleLocalize = map[uint16]struct {
		Rule1 map[string]map[string][3]interface{}
		Rule2 map[string][3]interface{}
	}

	ValidationExtraInfo struct {
		Key      string
		Required bool
		Validate func(value interface{}) bool
	}

	IFieldValidator interface {
		ValidFieldRules() ValidFieldResult
	}

	IExtraValidator interface {
		ValidExtraRules() (utils.KSMap, ValidExtraResult)
	}

	IStructValidator interface {
		ValidStructRules(obj any, fn ValidReportError)
	}

	IRuleLocalizes interface {
		ValidRuleLocalizes() ValidRuleLocalize
	}

	Validator struct {
		validate *validator.Validate
		any
	}
)

func NewValidator(obj any) *Validator {
	return &Validator{
		validate: valid.Get(),
		any:      obj,
	}
}

func (v *Validator) Valid(scene uint16) *err.CodeErrs {
	if !regTypes[reflect.TypeOf(v.any)] {
		// fields
		if i, ok := v.any.(IFieldValidator); ok {
			scenes := i.ValidFieldRules()
			sceneRules := make(map[string]func(reflect.Value) bool)
			if allRules := scenes[ValidSceneAll]; allRules != nil {
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
				sceneRules := make(map[string]ValidationExtraInfo)
				if allRules := scenes[ValidSceneAll]; allRules != nil {
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
				if allRules := scenes[ValidSceneAll]; allRules.Rule1 != nil {
					for name, rule := range allRules.Rule1 {
						sceneRule1s[name] = rule
					}
				}
				if allRules := scenes[ValidSceneAll]; allRules.Rule2 != nil {
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
										params = append(params, vvv[2].([]any))
									}
									if vvv[1].(bool) {
										params = append(params, ee.Param())
									}
									_ = cErrs.WrapLocalize(fmt.Sprintf(vvv[0].(string), params), nil)
								}
							}
						}
					}
					for kk, vv := range sceneRule2s {
						if ee.Tag() == kk {
							var params []any
							if vv[2] != nil {
								params = append(params, vv[2].([]any))
							}
							if vv[1].(bool) {
								params = append(params, ee.Param())
							}
							_ = cErrs.WrapLocalize(fmt.Sprintf(vv[0].(string), params), nil)
						}
					}
				}
			}
			return cErrs
		}
	}
	return err.Match(e)
}
