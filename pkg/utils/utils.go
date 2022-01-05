package utils

import "strings"

func DefaultStringIfEmpty(str string, defaultString string) string {
	if str != "" {
		return str
	}
	return defaultString
}

type SlugifyOptions struct {
	KeepHyphen bool
	KeepColon  bool
	KeepHash   bool
}

func Slugify(str string, options SlugifyOptions) string {
	str = strings.ToLower(str)
	str = strings.ReplaceAll(str, " ", "_")
	if !options.KeepHyphen {
		str = strings.ReplaceAll(str, "-", "_")
	}
	if !options.KeepColon {
		str = strings.ReplaceAll(str, ":", "_")
	}
	if !options.KeepHash {
		str = strings.ReplaceAll(str, "#", "_")
	}
	return str
}
