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

// TODO:GG 重复注册?

type (
	// ReportError validator.StructLevel.ReportError
	ReportError = func(field interface{}, fieldName, structFieldName string, tag, param string)

	// ExtraValidationRule 额外验证规则
	ExtraValidationRule struct {
		Key      string
		Required bool
		Validate func(value interface{}) bool
	}

	IFieldValidator interface {
		FieldRules() map[string]func(reflect.Value) bool
	}

	IExtraValidator interface {
		ExtraRules() (utils.KSMap, map[string]ExtraValidationRule)
	}

	IStructValidator interface {
		StructRules(obj any, fn ReportError)
	}

	IRuleLocalizes interface {
		RuleLocalizes() (map[string]map[string]string, map[string]string)
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

func (v *Validator) Valid() *err.CodeErrs {
	// fields
	if i, ok := v.any.(IFieldValidator); ok {
		fRules := i.FieldRules()
		for name, rule := range fRules {
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
			extra, rules := i.ExtraRules()
			// 验证Extra字段
			for key, rule := range rules {
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
			i.StructRules(sl.Current().Interface(), sl.ReportError)
		}, v)
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
				commonTags, customTags := i.RuleLocalizes()
				for _, ee := range validateErrs {
					for kk, vv := range commonTags {
						if ee.Tag() == kk {
							for kkk, vvv := range vv {
								if ee.Field() == kkk {
									_ = cErrs.WrapLocalize(fmt.Sprintf(vvv, ee.Param()), nil)
								}
							}
						}
					}
					for kk, vv := range customTags {
						if ee.Tag() == kk {
							_ = cErrs.WrapLocalize(fmt.Sprintf(vv, ee.Param()), nil)
						}
					}
				}
			}
			return cErrs
		}
	}
	return err.Match(e)
}
