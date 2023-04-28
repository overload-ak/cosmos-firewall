package utils

import (
	"regexp"
)

const (
	VariableStarPattern = `\{[a-zA-Z0-9_]+=*\*\}`
	VariablePattern     = `\{[a-zA-Z0-9_]+\}`
)

type PathPattern struct {
	Pattern string
	re      *regexp.Regexp
}

func NewPathPattern(pattern string) PathPattern {
	regexPattern := convertToRegexPattern(pattern)
	return PathPattern{
		Pattern: pattern,
		re:      regexp.MustCompile(regexPattern),
	}
}

func (p PathPattern) Match(url string) bool {
	return p.re.MatchString(url)
}

func convertToRegexPattern(pattern string) string {
	re := regexp.MustCompile(VariableStarPattern)
	regexPattern := re.ReplaceAllString(pattern, "(.*)")
	re = regexp.MustCompile(VariablePattern)
	regexPattern = re.ReplaceAllString(regexPattern, "([^/]+)")
	regexPattern = "^" + regexPattern + `(\?.*)?$`
	return regexPattern
}
