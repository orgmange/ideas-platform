package usecase

import (
	"regexp"
	"strings"
)

func validatePhone(phone string) bool {
	re := regexp.MustCompile(`^(\+7|8)\d{10}$`)
	return re.MatchString(phone)
}

func normalizePhone(phone string) string {
	if strings.HasPrefix(phone, "+7") {
		return phone[2:]
	}
	if strings.HasPrefix(phone, "8") {
		return phone[1:]
	}
	return phone
}