/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"strconv"
	"strings"
)

/*
Po parses the content of any PO file and provides all the Translation functions needed.
It's the base object used by all package methods.
And it's safe for concurrent use by multiple goroutines by using the sync package for locking.

Example:

	import (
		"fmt"
		"github.com/leonelquinteros/gotext"
	)

	func main() {
		// Create po object
		po := gotext.NewPoTranslator()

		// Parse .po file
		po.ParseFile("/path/to/po/file/translations.po")

		// Get Translation
		fmt.Println(po.Get("Translate this"))
	}

*/
type Po struct {
	//these three public members are for backwards compatibility. they are just set to the value in the domain
	Headers     HeaderMap
	Language    string
	PluralForms string

	domain *Domain
}

type parseState int

const (
	head parseState = iota
	msgCtxt
	msgID
	msgIDPlural
	msgStr
)

//NewPo should always be used to instantiate a new Po object
func NewPo() *Po {
	po := new(Po)
	po.domain = NewDomain()

	return po
}

func (po *Po) GetDomain() *Domain {
	return po.domain
}

// Convenience interfaces
func (po *Po) DropStaleTranslations() {
	po.domain.DropStaleTranslations()
}

func (po *Po) SetRefs(str string, refs []string) {
	po.domain.SetRefs(str, refs)
}
func (po *Po) GetRefs(str string) []string {
	return po.domain.GetRefs(str)
}

func (po *Po) Set(id, str string) {
	po.domain.Set(id, str)
}
func (po *Po) Get(str string, vars ...interface{}) string {
	return po.domain.Get(str, vars...)
}

func (po *Po) SetN(id, plural string, n int, str string) {
	po.domain.SetN(id, plural, n, str)
}
func (po *Po) GetN(str, plural string, n int, vars ...interface{}) string {
	return po.domain.GetN(str, plural, n, vars...)
}

func (po *Po) SetC(id, ctx, str string) {
	po.domain.SetC(id, ctx, str)
}
func (po *Po) GetC(str, ctx string, vars ...interface{}) string {
	return po.domain.GetC(str, ctx, vars...)
}

func (po *Po) SetNC(id, plural, ctx string, n int, str string) {
	po.domain.SetNC(id, plural, ctx, n, str)
}
func (po *Po) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return po.domain.GetNC(str, plural, n, ctx, vars...)
}

func (po *Po) MarshalText() ([]byte, error) {
	return po.domain.MarshalText()
}

func (po *Po) MarshalBinary() ([]byte, error) {
	return po.domain.MarshalBinary()
}

func (po *Po) UnmarshalBinary(data []byte) error {
	return po.domain.UnmarshalBinary(data)
}

func (po *Po) ParseFile(f string) {
	data, err := getFileData(f)
	if err != nil {
		return
	}

	po.Parse(data)
}

// Parse loads the translations specified in the provided string (str)
func (po *Po) Parse(buf []byte) {
	if po.domain == nil {
		panic("NewPo() was not used to instantiate this object")
	}

	// Lock while parsing
	po.domain.trMutex.Lock()
	po.domain.pluralMutex.Lock()
	defer po.domain.trMutex.Unlock()
	defer po.domain.pluralMutex.Unlock()

	// Get lines
	lines := strings.Split(string(buf), "\n")

	// Init buffer
	po.domain.trBuffer = NewTranslation()
	po.domain.ctxBuffer = ""
	po.domain.refBuffer = ""

	state := head
	for _, l := range lines {
		// Trim spaces
		l = strings.TrimSpace(l)

		// Skip invalid lines
		if !po.isValidLine(l) {
			po.parseComment(l, state)
			continue
		}

		// Buffer context and continue
		if strings.HasPrefix(l, "msgctxt") {
			po.parseContext(l)
			state = msgCtxt
			continue
		}

		// Buffer msgid and continue
		if strings.HasPrefix(l, "msgid") && !strings.HasPrefix(l, "msgid_plural") {
			po.parseID(l)
			state = msgID
			continue
		}

		// Check for plural form
		if strings.HasPrefix(l, "msgid_plural") {
			po.parsePluralID(l)
			po.domain.pluralTranslations[po.domain.trBuffer.PluralID] = po.domain.trBuffer
			state = msgIDPlural
			continue
		}

		// Save Translation
		if strings.HasPrefix(l, "msgstr") {
			po.parseMessage(l)
			state = msgStr
			continue
		}

		// Multi line strings and headers
		if strings.HasPrefix(l, "\"") && strings.HasSuffix(l, "\"") {
			po.parseString(l, state)
			continue
		}
	}

	// Save last Translation buffer.
	po.saveBuffer()

	// Parse headers
	po.domain.parseHeaders()

	// set values on this struct
	// this is for backwards compatibility
	po.Language = po.domain.Language
	po.PluralForms = po.domain.PluralForms
	po.Headers = po.domain.Headers
}

// saveBuffer takes the context and Translation buffers
// and saves it on the translations collection
func (po *Po) saveBuffer() {
	// With no context...
	if po.domain.ctxBuffer == "" {
		po.domain.translations[po.domain.trBuffer.ID] = po.domain.trBuffer
	} else {
		// With context...
		if _, ok := po.domain.contexts[po.domain.ctxBuffer]; !ok {
			po.domain.contexts[po.domain.ctxBuffer] = make(map[string]*Translation)
		}
		po.domain.contexts[po.domain.ctxBuffer][po.domain.trBuffer.ID] = po.domain.trBuffer

		// Cleanup current context buffer if needed
		if po.domain.trBuffer.ID != "" {
			po.domain.ctxBuffer = ""
		}
	}

	// Flush Translation buffer
	if po.domain.refBuffer == "" {
		po.domain.trBuffer = NewTranslation()
	} else {
		po.domain.trBuffer = NewTranslationWithRefs(strings.Split(po.domain.refBuffer, " "))
	}
}

// Either preserves comments before the first "msgid", for later round-trip.
// Or preserves source references for a given translation.
func (po *Po) parseComment(l string, state parseState) {
	if len(l) > 0 && l[0] == '#' {
		if state == head {
			po.domain.headerComments = append(po.domain.headerComments, l)
		} else if len(l) > 1 {
			switch l[1] {
			case ':':
				if len(l) > 2 {
					po.domain.refBuffer = strings.TrimSpace(l[2:])
				}
			}
		}
	}
}

// parseContext takes a line starting with "msgctxt",
// saves the current Translation buffer and creates a new context.
func (po *Po) parseContext(l string) {
	// Save current Translation buffer.
	po.saveBuffer()

	// Buffer context
	po.domain.ctxBuffer, _ = strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(l, "msgctxt")))
}

// parseID takes a line starting with "msgid",
// saves the current Translation and creates a new msgid buffer.
func (po *Po) parseID(l string) {
	// Save current Translation buffer.
	po.saveBuffer()

	// Set id
	po.domain.trBuffer.ID, _ = strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(l, "msgid")))
}

// parsePluralID saves the plural id buffer from a line starting with "msgid_plural"
func (po *Po) parsePluralID(l string) {
	po.domain.trBuffer.PluralID, _ = strconv.Unquote(strings.TrimSpace(strings.TrimPrefix(l, "msgid_plural")))
}

// parseMessage takes a line starting with "msgstr" and saves it into the current buffer.
func (po *Po) parseMessage(l string) {
	l = strings.TrimSpace(strings.TrimPrefix(l, "msgstr"))

	// Check for indexed Translation forms
	if strings.HasPrefix(l, "[") {
		idx := strings.Index(l, "]")
		if idx == -1 {
			// Skip wrong index formatting
			return
		}

		// Parse index
		i, err := strconv.Atoi(l[1:idx])
		if err != nil {
			// Skip wrong index formatting
			return
		}

		// Parse Translation string
		po.domain.trBuffer.Trs[i], _ = strconv.Unquote(strings.TrimSpace(l[idx+1:]))

		// Loop
		return
	}

	// Save single Translation form under 0 index
	po.domain.trBuffer.Trs[0], _ = strconv.Unquote(l)
}

// parseString takes a well formatted string without prefix
// and creates headers or attach multi-line strings when corresponding
func (po *Po) parseString(l string, state parseState) {
	clean, _ := strconv.Unquote(l)

	switch state {
	case msgStr:
		// Append to last Translation found
		po.domain.trBuffer.Trs[len(po.domain.trBuffer.Trs)-1] += clean

	case msgID:
		// Multiline msgid - Append to current id
		po.domain.trBuffer.ID += clean

	case msgIDPlural:
		// Multiline msgid - Append to current id
		po.domain.trBuffer.PluralID += clean

	case msgCtxt:
		// Multiline context - Append to current context
		po.domain.ctxBuffer += clean

	}
}

// isValidLine checks for line prefixes to detect valid syntax.
func (po *Po) isValidLine(l string) bool {
	// Check prefix
	valid := []string{
		"\"",
		"msgctxt",
		"msgid",
		"msgid_plural",
		"msgstr",
	}

	for _, v := range valid {
		if strings.HasPrefix(l, v) {
			return true
		}
	}

	return false
}
