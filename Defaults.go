package vutils

import "strings"

type defaultsUtils struct{}

func (du *defaultsUtils) DefaultBoolFromString(s2 string) bool {

	if strings.ToLower(s2) == "true" {
		return true
	}
	return false
}

func (du *defaultsUtils) DefaultBool(s1 bool, s2 bool) bool {

	if !s1 && s2 {
		return s2
	}
	return s1
}

var Defaults = &defaultsUtils{}
