package utils

import (
	"regexp"
	"strings"
)

func Slugify(name string) string {
	slug := strings.ToLower(name)

	slug=strings.ReplaceAll(slug," ","-")

	reg:=regexp.MustCompile(`[^a-z0-9-]`)

	slug=reg.ReplaceAllString(slug,"")

	return slug
}