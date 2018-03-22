package gitattributes

import (
	"path/filepath"
	"strings"
)

func unjoin(qualified, base string) string {
	return strings.TrimPrefix(filepath.ToSlash(qualified), filepath.ToSlash(base))
}

func join(parts ...string) string {
	var path string
	for i, part := range parts {
		part = filepath.ToSlash(part)
		if i > 0 {
			part = strings.TrimPrefix(part, "/")
		}
		part = strings.TrimSuffix(part, "/")

		path = path + part
		if i+1 < len(parts) {
			path = path + "/"
		}
	}

	return path
}

func rooted(path string) bool {
	return strings.HasPrefix(filepath.ToSlash(path), "/")
}

var escapes = map[string]string{
	" ": "[[:space:]]",
	"#": "\\#",
}

func escape(s string) string {
	s = filepath.ToSlash(s)

	for from, to := range escapes {
		s = strings.Replace(s, from, to, -1)
	}
	return s
}

func unescape(s string) string {
	for to, from := range escapes {
		s = strings.Replace(s, from, to, -1)
	}
	return s
}
