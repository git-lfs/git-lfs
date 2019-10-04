// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// cookie parses curl cookie jar files.
package cookie

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ParseCookieJar parses a cURL/Mozilla/Netscape cookie jar text file.
func ParseCookieJar(r io.Reader) ([]*http.Cookie, error) {
	var result []*http.Cookie
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		httpOnly := false
		const httpOnlyPrefix = "#HttpOnly_"
		if strings.HasPrefix(line, httpOnlyPrefix) {
			line = line[len(httpOnlyPrefix):]
			httpOnly = true
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 6 && len(fields) != 7 {
			return nil, fmt.Errorf("got %d fields in line %q, want 6 or 7", len(fields), line)
		}

		exp, err := strconv.ParseInt(fields[4], 10, 64)
		if err != nil {
			return nil, err
		}

		c := http.Cookie{
			Domain:   strings.TrimSpace(fields[0]),
			Name:     strings.TrimSpace(fields[5]),
			Path:     strings.TrimSpace(fields[2]),
			Expires:  time.Unix(exp, 0),
			Secure:   fields[3] == "TRUE",
			HttpOnly: httpOnly,
		}
		if len(fields) == 7 {
			c.Value = strings.TrimSpace(fields[6])
		}

		result = append(result, &c)
	}

	return result, nil
}

// WatchJar starts watching the given path for changes, and loads new
// data from the file whenever it is available.
func WatchJar(jar http.CookieJar, path string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// We watch the dir, so we catch creation + rename events too.
	if err := w.Add(filepath.Dir(path)); err != nil {
		return err
	}

	go func() {
		var lastMod time.Time
		for {
			select {
			case <-w.Events:
				fi, err := os.Stat(path)
				if os.IsNotExist(err) {
					continue
				}

				if err != nil {
					log.Printf("Stat(%s): %v", path, err)
					continue
				}
				if fi.ModTime().Equal(lastMod) {
					continue
				}
				lastMod = fi.ModTime()
				if err := updateJar(jar, path); err != nil {
					log.Printf("updateJar(%s): %v", path, err)
				}

			case <-w.Errors:
				log.Printf("notify (%s):  %v", path, err)
			}
		}
	}()
	return nil
}

// updateJar reads the path into the jar.
func updateJar(jar http.CookieJar, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	cs, err := ParseCookieJar(f)
	if err != nil {
		return err
	}

	for _, c := range cs {
		jar.SetCookies(&url.URL{
			Scheme: "http",
			Host:   c.Domain,
		}, []*http.Cookie{c})
	}

	return nil
}

// NewJar reads cookies in the Mozilla/Netscape cookie file format,
// and returns them as a CookieJar
func NewJar(path string) (http.CookieJar, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	if err := updateJar(jar, path); err != nil {
		return nil, err
	}

	return jar, nil
}
