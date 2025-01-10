package validates

import (
	"Gin-IM/pkg/defines"
	"Gin-IM/pkg/exception"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"reflect"
)

func newValidator() *validator.Validate {
	valid := validator.New()
	for key, value := range rules {
		if err := valid.RegisterValidation(key, value); err != nil {
			log.Logger.Fatal().Err(err).Msg("register validation error")
		}
	}
	return valid
}

func processError(data interface{}, err error) string {
	if err == nil {
		return ""
	}
	var invalid *validator.InvalidValidationError
	if errors.As(err, &invalid) {
		return fmt.Sprintf("输入参数错误: %s", invalid.Error())
	}
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, validationError := range validationErrs {
			fieldName := validationError.Field()
			typeOf := reflect.TypeOf(data)
			if typeOf.Kind() == reflect.Pointer {
				typeOf = typeOf.Elem()
			}
			if field, o := typeOf.FieldByName(fieldName); o {
				errorInfo := field.Tag.Get(defines.FIELD_ERROR_INFO)
				return fmt.Sprintf("%s : %s", fieldName, errorInfo)
			} else {
				return "缺失字段错误信息"
			}
		}
	}
	return ""
}

func Validate(data interface{}) error {
	if errs := newValidator().Struct(data); errs != nil {
		return exception.NewError(1005, processError(data, errs))
	}
	return nil
}
