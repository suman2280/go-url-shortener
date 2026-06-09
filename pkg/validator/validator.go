package validator

import (
	"net/url"
	"regexp"
)

var aliasRegex = regexp.MustCompile(`^[a-zA-Z0-9-]+$`)

func IsValidAlias(alias string) bool {
	return len(alias) >= 1 && len(alias) <= 50 && aliasRegex.MatchString(alias)
}

func IsValidURL(u string) bool {
	parsed, err := url.ParseRequestURI(u)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host != ""
}
