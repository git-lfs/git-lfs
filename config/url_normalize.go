package config

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	urlUnsafeChars   = " <>\"%{}|\\^`"
	urlReservedChars = ":/?#[]@!$&'()*+,;="
)

type normalizedURL struct {
	scheme      string
	userinfo    string
	hasUserinfo bool
	username    string
	host        string
	hostname    string
	port        string
	path        string
	suffix      string
}

func (u *normalizedURL) String() string {
	var b strings.Builder
	b.WriteString(u.scheme)
	b.WriteString("://")
	if u.hasUserinfo {
		b.WriteString(u.userinfo)
		b.WriteByte('@')
	}
	b.WriteString(u.host)
	if u.port != "" {
		b.WriteByte(':')
		b.WriteString(u.port)
	}
	b.WriteString(u.path)
	b.WriteString(u.suffix)
	return b.String()
}

// normalizeURL normalizes a complete URL using the matching rules Git applies
// to URL-scoped configuration. It intentionally keeps reserved characters
// escaped when they were escaped in the input, since decoding them could alter
// URL component boundaries.
func normalizeURL(rawurl string, allowGlobs bool) (*normalizedURL, error) {
	schemeEnd := strings.Index(rawurl, "://")
	if schemeEnd <= 0 || !validURLScheme(rawurl[:schemeEnd]) {
		return nil, fmt.Errorf("invalid URL scheme or missing ://")
	}

	scheme := strings.ToLower(rawurl[:schemeEnd])
	remainder := rawurl[schemeEnd+3:]
	authorityEnd := strings.IndexAny(remainder, "/?#")
	if authorityEnd < 0 {
		authorityEnd = len(remainder)
	}
	authority := remainder[:authorityEnd]
	tail := remainder[authorityEnd:]

	hasUserinfo := false
	userinfo := ""
	username := ""
	if at := strings.IndexByte(authority, '@'); at >= 0 {
		hasUserinfo = true
		var err error
		userinfo, err = normalizeURLEscapes(authority[:at])
		if err != nil {
			return nil, err
		}
		username = userinfo
		if colon := strings.IndexByte(username, ':'); colon >= 0 {
			username = username[:colon]
		}
		authority = authority[at+1:]
	}

	host, hostname, rawPort, hasPort, err := splitURLHostPort(authority)
	if err != nil {
		return nil, err
	}
	if hostname == "" && scheme != "file" {
		return nil, fmt.Errorf("missing host")
	}
	if !validURLHost(hostname, allowGlobs) {
		return nil, fmt.Errorf("invalid characters in host name")
	}

	host = strings.ToLower(host)
	hostname = strings.ToLower(hostname)
	port := ""
	if hasPort && rawPort != "" {
		if scheme == "file" {
			return nil, fmt.Errorf("file URL may not have a port")
		}
		if !allASCIIDigits(rawPort) {
			return nil, fmt.Errorf("invalid port number")
		}
		rawPort = strings.TrimLeft(rawPort, "0")
		if rawPort == "" {
			return nil, fmt.Errorf("invalid port number")
		}
		portNumber, err := strconv.Atoi(rawPort)
		if err != nil || portNumber < 1 || portNumber > 65535 {
			return nil, fmt.Errorf("invalid port number")
		}
		if !((scheme == "http" && portNumber == 80) ||
			(scheme == "https" && portNumber == 443) ||
			(scheme == "ssh" && portNumber == 22)) {
			port = strconv.Itoa(portNumber)
		}
	}

	pathEnd := strings.IndexAny(tail, "?#")
	if pathEnd < 0 {
		pathEnd = len(tail)
	}
	path, err := normalizeURLPath(tail[:pathEnd])
	if err != nil {
		return nil, err
	}
	suffix, err := normalizeURLEscapes(tail[pathEnd:])
	if err != nil {
		return nil, err
	}

	return &normalizedURL{
		scheme:      scheme,
		userinfo:    userinfo,
		hasUserinfo: hasUserinfo,
		username:    username,
		host:        host,
		hostname:    hostname,
		port:        port,
		path:        path,
		suffix:      suffix,
	}, nil
}

func validURLScheme(scheme string) bool {
	for i := 0; i < len(scheme); i++ {
		c := scheme[i]
		if i == 0 && !isASCIIAlpha(c) {
			return false
		}
		if !isASCIIAlpha(c) && !isASCIIDigit(c) && c != '+' && c != '-' && c != '.' {
			return false
		}
	}
	return scheme != ""
}

func splitURLHostPort(authority string) (host, hostname, port string, hasPort bool, err error) {
	if strings.HasPrefix(authority, "[") {
		closeBracket := strings.IndexByte(authority, ']')
		if closeBracket < 0 {
			return "", "", "", false, fmt.Errorf("invalid IPv6 host")
		}
		host = authority[:closeBracket+1]
		hostname = authority[1:closeBracket]
		remainder := authority[closeBracket+1:]
		if remainder == "" {
			return host, hostname, "", false, nil
		}
		if remainder[0] != ':' {
			return "", "", "", false, fmt.Errorf("invalid host or port")
		}
		return host, hostname, remainder[1:], true, nil
	}

	colon := strings.LastIndexByte(authority, ':')
	if colon >= 0 {
		if strings.Contains(authority[:colon], ":") {
			return "", "", "", false, fmt.Errorf("IPv6 host must be bracketed")
		}
		return authority[:colon], authority[:colon], authority[colon+1:], true, nil
	}
	return authority, authority, "", false, nil
}

func validURLHost(host string, allowGlobs bool) bool {
	for i := 0; i < len(host); i++ {
		c := host[i]
		if isASCIIAlpha(c) || isASCIIDigit(c) || strings.ContainsRune(".-_:", rune(c)) {
			continue
		}
		if allowGlobs && c == '*' {
			continue
		}
		return false
	}
	return true
}

func normalizeURLPath(rawPath string) (string, error) {
	if strings.HasPrefix(rawPath, "/") {
		rawPath = rawPath[1:]
	}

	path := []byte{'/'}
	for {
		nextSlash := strings.IndexByte(rawPath, '/')
		rawSegment := rawPath
		if nextSlash >= 0 {
			rawSegment = rawPath[:nextSlash]
		}

		segment, err := normalizeURLEscapes(rawSegment)
		if err != nil {
			return "", err
		}
		segmentStart := len(path)
		path = append(path, segment...)
		skipSlash := false

		switch segment {
		case ".":
			if segmentStart == 1 {
				path = path[:len(path)-1]
				skipSlash = true
			} else {
				path = path[:len(path)-2]
			}
		case "..":
			previousSlash := len(path) - 3
			if previousSlash == 0 {
				return "", fmt.Errorf("invalid .. path segment")
			}
			for {
				previousSlash--
				if previousSlash < 0 {
					return "", fmt.Errorf("invalid .. path segment")
				}
				if path[previousSlash] == '/' {
					break
				}
			}
			if previousSlash == 0 {
				path = path[:1]
				skipSlash = true
			} else {
				path = path[:previousSlash]
			}
		}

		if nextSlash < 0 {
			break
		}
		rawPath = rawPath[nextSlash+1:]
		if !skipSlash {
			path = append(path, '/')
		}
	}

	return string(path), nil
}

func normalizeURLEscapes(raw string) (string, error) {
	var b strings.Builder
	b.Grow(len(raw))

	for i := 0; i < len(raw); i++ {
		c := raw[i]
		wasEscaped := false
		if c == '%' {
			if i+2 >= len(raw) {
				return "", fmt.Errorf("invalid URL escape")
			}
			high, okHigh := fromHex(raw[i+1])
			low, okLow := fromHex(raw[i+2])
			if !okHigh || !okLow {
				return "", fmt.Errorf("invalid URL escape")
			}
			c = high<<4 | low
			i += 2
			wasEscaped = true
		}

		mustEscape := c <= 0x1f || c >= 0x7f ||
			strings.ContainsRune(urlUnsafeChars, rune(c)) ||
			(wasEscaped && strings.ContainsRune(urlReservedChars, rune(c)))
		if mustEscape {
			const hex = "0123456789ABCDEF"
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		} else {
			b.WriteByte(c)
		}
	}

	return b.String(), nil
}

func fromHex(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

func allASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if !isASCIIDigit(value[i]) {
			return false
		}
	}
	return true
}

func isASCIIAlpha(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z'
}

func isASCIIDigit(c byte) bool {
	return c >= '0' && c <= '9'
}
