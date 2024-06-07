package abi

import "regexp"

var fixedStringRegexCheck = regexp.MustCompile(".*\\[[0-9]+\\]")
var dynamicRegexCheck = regexp.MustCompile("([a-z]+)\\[[0-9]+\\]")

func regexFixedString(s string) bool {
	return fixedStringRegexCheck.MatchString(s)
}

func regexDynamic(s string) bool {
	return dynamicRegexCheck.MatchString(s)
}
