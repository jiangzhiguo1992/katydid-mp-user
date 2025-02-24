package model

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"katydid-mp-user/pkg/valid"
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

	ValidFieldError struct {
		validator.FieldError
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

	IRuleLocalizes interface {
		RuleLocalizes(err ValidFieldError)
	}

	Validator struct {
		validate *validator.Validate
		any
	}
)

func (ve *ValidFieldError) Tag() string {
	return ve.FieldError.Tag()
}

func (ve *ValidFieldError) Field() string {
	return ve.FieldError.Field()
}

func (ve *ValidFieldError) Error() string {
	return ve.FieldError.Error()
}

func NewValidator(obj any) *Validator {
	return &Validator{
		validate: valid.Get(),
		any:      obj,
	}
}

func (v *Validator) Valid() error {
	// fields
	if i, ok := v.any.(IFieldValidator); ok {
		fRules := i.FieldRules()
		for name, rule := range fRules {
			e := v.validate.RegisterValidation(name, func(fl validator.FieldLevel) bool {
				return rule(fl.Field())
			})
			if e != nil {
				return e
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
	//if i, ok := v.any.(IRuleLocalizes); ok {
	//	// 注册错误消息
	//	v.validate.RegisterTranslation("client-name", nil,
	//		func(ut ut.Translator) error {
	//			return ut.Add("client-name", "{0}格式不正确", true)
	//		},
	//		func(ut ut.Translator, fe validator.FieldError) string {
	//			t, _ := ut.T("client-name", fe.Field())
	//			return t
	//		},
	//	)
	//}

	return v.validate.Struct(v)
}
