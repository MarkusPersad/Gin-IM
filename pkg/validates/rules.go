package validates

import (
	"Gin-IM/pkg/defines"
	"github.com/dlclark/regexp2"
	"github.com/go-playground/validator/v10"
)

var rules = map[string]func(fl validator.FieldLevel) bool{}

func init() {
	rules["pass"] = checkPassword
}
func checkPassword(fl validator.FieldLevel) bool {
	re := regexp2.MustCompile(defines.PASSWORD_REGEX, 0)
	if isMatch, _ := re.MatchString(fl.Field().String()); isMatch {
		return true
	}
	return false
}
