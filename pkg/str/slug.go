package str

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
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
in case user email begins with a digit, 20230123@acme.org the returned user slug is `u2023123_acme_org`,
otherwise removes all the non-alpha numeric charecters from the provided email and returns an _ seperated slug.
For eg: "$john-doe@acme.org" returns "johndoe_acme_org"
*/
func GenerateUserSlug(email string) string {
	email = strings.ToLower(strings.TrimSpace(email))

	i := strings.LastIndexByte(email, '@')
	// remove all the non-alphanumeric charecters from local part
	regex := "[^a-zA-Z0-9]+"
	localPart := regexp.MustCompile(regex).ReplaceAllString(email[:i], "")
	if unicode.IsDigit(rune(localPart[0])) {
		localPart = fmt.Sprintf("u%s", localPart)
	}

	// remove all the non-numeric charecters except periods from the domain part
	regex = "[^a-zA-Z0-9.]+"
	domainPart := strings.ReplaceAll(regexp.MustCompile(regex).ReplaceAllString(email[i+1:], ""), ".", "_")

	return fmt.Sprintf("%s_%s", localPart, domainPart)
}
