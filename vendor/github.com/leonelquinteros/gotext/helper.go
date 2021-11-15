/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"fmt"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`%\(([a-zA-Z0-9_]+)\)[.0-9]*[svTtbcdoqXxUeEfFgGp]`)

// SimplifiedLocale simplified locale like " en_US"/"de_DE "/en_US.UTF-8/zh_CN/zh_TW/el_GR@euro/... to en_US, de_DE, zh_CN, el_GR...
func SimplifiedLocale(lang string) string {
	// en_US/en_US.UTF-8/zh_CN/zh_TW/el_GR@euro/...
	if idx := strings.Index(lang, ":"); idx != -1 {
		lang = lang[:idx]
	}
	if idx := strings.Index(lang, "@"); idx != -1 {
		lang = lang[:idx]
	}
	if idx := strings.Index(lang, "."); idx != -1 {
		lang = lang[:idx]
	}
	return strings.TrimSpace(lang)
}

// Printf applies text formatting only when needed to parse variables.
func Printf(str string, vars ...interface{}) string {
	if len(vars) > 0 {
		return fmt.Sprintf(str, vars...)
	}

	return str
}

// NPrintf support named format
// NPrintf("%(name)s is Type %(type)s", map[string]interface{}{"name": "Gotext", "type": "struct"})
func NPrintf(format string, params map[string]interface{}) {
	f, p := parseSprintf(format, params)
	fmt.Printf(f, p...)
}

// Sprintf support named format
//      Sprintf("%(name)s is Type %(type)s", map[string]interface{}{"name": "Gotext", "type": "struct"})
func Sprintf(format string, params map[string]interface{}) string {
	f, p := parseSprintf(format, params)
	return fmt.Sprintf(f, p...)
}

func parseSprintf(format string, params map[string]interface{}) (string, []interface{}) {
	f, n := reformatSprintf(format)
	var p []interface{}
	for _, v := range n {
		p = append(p, params[v])
	}
	return f, p
}

func reformatSprintf(f string) (string, []string) {
	m := re.FindAllStringSubmatch(f, -1)
	i := re.FindAllStringSubmatchIndex(f, -1)

	ord := []string{}
	for _, v := range m {
		ord = append(ord, v[1])
	}

	pair := []int{0}
	for _, v := range i {
		pair = append(pair, v[2]-1)
		pair = append(pair, v[3]+1)
	}
	pair = append(pair, len(f))
	plen := len(pair)

	out := ""
	for n := 0; n < plen; n += 2 {
		out += f[pair[n]:pair[n+1]]
	}

	return out, ord
}
