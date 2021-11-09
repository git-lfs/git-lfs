package tr

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/leonelquinteros/gotext"
)

//go:generate go run ../tr/trgen/trgen.go

var Tr = gotext.NewLocale("/usr/share/locale", "en")

var locales = make(map[string]string)

func findLocale() string {
	vars := []string{"LC_ALL", "LC_MESSAGES", "LANG"}
	for _, varname := range vars {
		if val, ok := os.LookupEnv(varname); ok {
			return val
		}
	}
	return ""
}

func processLocale(locale string) []string {
	options := make([]string, 0, 2)
	// For example, split "en_DK.UTF-8" into "en_DK" and "UTF-8".
	pieces := strings.Split(locale, ".")
	options = append(options, pieces[0])
	// For example, split "en_DK" into "en" and "DK".
	pieces = strings.Split(pieces[0], "_")
	if len(pieces) > 1 {
		options = append(options, pieces[0])
	}
	return options
}

func InitializeLocale() {
	locale := findLocale()
	if len(locale) == 0 {
		return
	}
	Tr = gotext.NewLocale("/usr/share/locale", locale)
	Tr.AddDomain("git-lfs")
	for _, loc := range processLocale(locale) {
		if moData, ok := locales[loc]; ok {
			mo := gotext.NewMo()
			decodedData, err := base64.StdEncoding.DecodeString(moData)
			if err != nil {
				continue
			}
			mo.Parse(decodedData)
			Tr.AddTranslator("git-lfs", mo)
			return
		}
	}
}
