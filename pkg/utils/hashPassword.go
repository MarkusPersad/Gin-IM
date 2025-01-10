package utils

import (
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"os"
	"strconv"
)

var cost int

func init() {
	if val, err := strconv.Atoi(os.Getenv("HASH_SALT")); err == nil {
		cost = val
	} else {
		cost = bcrypt.DefaultCost
	}
}

func GernerateHashPassword(password string) string {
	if hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), cost); err != nil {
		log.Logger.Error().Err(err).Msg("failed to generate hash password")
		return ""
	} else {
		return string(hashPassword)
	}
}

func CompareHashPassword(hashPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashPassword), []byte(password)) == nil
}
