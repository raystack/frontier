package str

import "strings"

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

func GenerateSlug(name string) string {
	preProcessed := strings.ReplaceAll(strings.TrimSpace(strings.TrimSpace(name)), "_", "-")
	return strings.Join(
		strings.Split(preProcessed, " "),
		"-",
	)
}
