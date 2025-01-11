package utils

import (
	"os"
	"strings"
)

var admins []string

func init() {
	adminString := os.Getenv("ADMIN")
	if strings.TrimSpace(adminString) != "" {
		if strings.Contains(adminString, ",") {
			admins = strings.Split(adminString, ",")
		}
		admins = append(admins, adminString)
	}
}

func IsAdmin(str string) bool {
	return contains(admins, str)
}
func contains[T comparable](collections []T, element T) bool {
	for _, value := range collections {
		if element == value {
			return true
		}
	}
	return false
}
