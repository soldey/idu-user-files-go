package common

import "strings"

func PrepareFilename(name string) (string, string) {
	preparedFilename := strings.Split(name, ".")
	filename := strings.Join(preparedFilename[:len(preparedFilename)-1], "")
	ext := preparedFilename[len(preparedFilename)-1]
	return filename, ext
}

type PlatformType string

const (
	DUP PlatformType = "DUP"
	TRP PlatformType = "TRP"
)
