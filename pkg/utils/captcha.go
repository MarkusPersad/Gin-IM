package utils

import (
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/types"
	"github.com/mojocn/base64Captcha"
	"image/color"
)

type Captcha struct {
	store  base64Captcha.Store
	driver base64Captcha.Driver
}

func NewCaptcha(store base64Captcha.Store) *Captcha {
	mathDriver := base64Captcha.NewDriverMath(40, 160, 5, base64Captcha.OptionShowSineLine, &color.RGBA{
		R: 254,
		G: 254,
		B: 254,
		A: 254,
	}, base64Captcha.DefaultEmbeddedFonts, []string{"wqy-microhei.ttc"})
	return &Captcha{
		store:  store,
		driver: mathDriver,
	}
}
func (c *Captcha) Generate() (*types.CaptDateBase64, error) {
	capt := base64Captcha.NewCaptcha(c.driver, c.store)
	id, b64s, _, err := capt.Generate()
	if err != nil {
		return nil, exception.ErrCheckCode
	}
	return &types.CaptDateBase64{
		Id:   id,
		B64s: b64s,
	}, nil
}

func (c *Captcha) Verify(id, answer string, clear bool) error {
	if len(id) == 0 || len(answer) == 0 {
		return exception.ErrCheckCode
	}
	if !base64Captcha.NewCaptcha(c.driver, c.store).Verify(id, answer, clear) {
		return exception.ErrCheckCode
	}
	return nil
}
