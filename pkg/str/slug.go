package str

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

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

/*
in case the user name doesnt begin with an alphabet this function returns a slug of type "user_1682290750".
Otherwise, slug of type "John_Doe_1682290750" is returned, with name being appended with epoch time from now
*/
func GenerateUserSlug(name string) string {
	regex := "^[a-zA-Z]"
	preProcessed := strings.TrimSpace(strings.TrimSpace(name))
	if !regexp.MustCompile(regex).MatchString(preProcessed) {
		preProcessed = "user"
	}
	epoch := time.Now().UTC().Unix()
	preProcessed = fmt.Sprintf("%s %v", preProcessed, epoch)
	return strings.Join(
		strings.Split(preProcessed, " "),
		"_",
	)
}
