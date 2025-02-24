package model

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
)

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
		ExtraRules() (map[string]any, map[string]ExtraValidationRule)
	}

	IStructValidator interface {
		StructRules(obj any, fn ReportError)
	}

	Validator struct {
		validate *validator.Validate
		any
	}
)

func NewValidator(validate *validator.Validate, obj any) *Validator {
	return &Validator{
		validate: validate,
		any:      obj,
	}
}

func (v *Validator) Valid() error {
	// fields
	if i, ok := v.any.(IFieldValidator); ok {
		fRules := i.FieldRules()
		for name, rule := range fRules {
			_ = v.validate.RegisterValidation(name, func(fl validator.FieldLevel) bool {
				return rule(fl.Field())
			})
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
	return v.validate.Struct(v)
}
