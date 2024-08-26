package common

import (
	"fmt"
	"strings"
)

func PrepareFilename(name string) (string, string) {
	preparedFilename := strings.Split(name, ".")
	filename := strings.Join(preparedFilename[:len(preparedFilename)-1], "")
	ext := preparedFilename[len(preparedFilename)-1]
	return filename, ext
}

func ParseParam(param string, isString bool) string {
	if param == "" {
		return "null"
	}
	if isString {
		return fmt.Sprintf("\"%s\"", param)
	}
	return param
}

type PlatformType string

const (
	DUP PlatformType = "DUP"
	TRP PlatformType = "TRP"
)
