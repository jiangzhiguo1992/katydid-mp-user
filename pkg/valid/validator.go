package valid

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"sync"
)

var (
	validate *validator.Validate
	once     sync.Once
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
