package soapboxd

import (
	"regexp"
	"strings"
)

var (
	slugSpaceRe      = regexp.MustCompile(`\s+`)
	slugNotAllowedRe = regexp.MustCompile(`[^a-z0-9-]`)
	slugRepeatDashRe = regexp.MustCompile(`-{2,}`)
)

func slugify(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	s = slugSpaceRe.ReplaceAllString(s, "-")
	s = slugNotAllowedRe.ReplaceAllString(s, "")
	s = slugRepeatDashRe.ReplaceAllString(s, "-")
	return s
}
