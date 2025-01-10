package database

import (
	"Gin-IM/pkg/defines"
	"context"
	"github.com/rs/zerolog/log"
)

var ctx = context.Background()

func (s *service) Set(id string, value string) error {
	key := defines.CAPTCHA + id
	return s.valClient.Do(ctx, s.valClient.B().Setex().Key(key).Seconds(defines.CAPTCHA_TIMEOUT).Value(value).Build()).Error()
}
func (s *service) Get(id string, clear bool) string {
	key := defines.CAPTCHA + id
	result := s.valClient.Do(ctx, s.valClient.B().Get().Key(key).Build())
	if result.Error() != nil {
		return ""
	}
	if clear {
		e := s.valClient.Do(ctx, s.valClient.B().Del().Key(key).Build()).Error()
		if e != nil {
			log.Logger.Error().Err(e).Msg("valkey del error")
			return ""
		}
	}
	val, e := result.ToString()
	if e != nil {
		log.Logger.Error().Err(e).Msg("valkey get error")
		return ""
	}
	return val
}

func (s *service) Verify(id, answer string, clear bool) bool {
	val := s.Get(id, clear)
	return val == answer
}
