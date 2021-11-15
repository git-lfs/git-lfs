package gotext

import (
	"bytes"
	"encoding/gob"
	"sort"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/text/language"

	"github.com/leonelquinteros/gotext/plurals"
)

// Domain has all the common functions for dealing with a gettext domain
// it's initialized with a GettextFile (which represents either a Po or Mo file)
type Domain struct {
	Headers HeaderMap

	// Language header
	Language string
	tag      language.Tag

	// Plural-Forms header
	PluralForms string

	// Preserve comments at head of PO for round-trip
	headerComments []string

	// Parsed Plural-Forms header values
	nplurals    int
	plural      string
	pluralforms plurals.Expression

	// Storage
	translations       map[string]*Translation
	contexts           map[string]map[string]*Translation
	pluralTranslations map[string]*Translation

	// Sync Mutex
	trMutex     sync.RWMutex
	pluralMutex sync.RWMutex

	// Parsing buffers
	trBuffer  *Translation
	ctxBuffer string
	refBuffer string
}

// Preserve MIMEHeader behaviour, without the canonicalisation
type HeaderMap map[string][]string

func (m HeaderMap) Add(key, value string) {
	m[key] = append(m[key], value)
}
func (m HeaderMap) Del(key string) {
	delete(m, key)
}
func (m HeaderMap) Get(key string) string {
	if m == nil {
		return ""
	}
	v := m[key]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}
func (m HeaderMap) Set(key, value string) {
	m[key] = []string{value}
}
func (m HeaderMap) Values(key string) []string {
	if m == nil {
		return nil
	}
	return m[key]
}

func NewDomain() *Domain {
	domain := new(Domain)

	domain.Headers = make(HeaderMap)
	domain.headerComments = make([]string, 0)
	domain.translations = make(map[string]*Translation)
	domain.contexts = make(map[string]map[string]*Translation)
	domain.pluralTranslations = make(map[string]*Translation)

	return domain
}

func (do *Domain) pluralForm(n int) int {
	// do we really need locking here? not sure how this plurals.Expression works, so sticking with it for now
	do.pluralMutex.RLock()
	defer do.pluralMutex.RUnlock()

	// Failure fallback
	if do.pluralforms == nil {
		/* Use the Germanic plural rule.  */
		if n == 1 {
			return 0
		}
		return 1
	}
	return do.pluralforms.Eval(uint32(n))
}

// parseHeaders retrieves data from previously parsed headers. it's called by both Mo and Po when parsing
func (do *Domain) parseHeaders() {
	raw := ""
	if _, ok := do.translations[raw]; ok {
		raw = do.translations[raw].Get()
	}

	// textproto.ReadMIMEHeader() forces keys through CanonicalMIMEHeaderKey(); must read header manually to have one-to-one round-trip of keys
	languageKey := "Language"
	pluralFormsKey := "Plural-Forms"

	rawLines := strings.Split(raw, "\n")
	for _, line := range rawLines {
		if len(line) == 0 {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}

		key := line[:colonIdx]
		lowerKey := strings.ToLower(key)
		if lowerKey == strings.ToLower(languageKey) {
			languageKey = key
		} else if lowerKey == strings.ToLower(pluralFormsKey) {
			pluralFormsKey = key
		}

		value := strings.TrimSpace(line[colonIdx+1:])
		do.Headers.Add(key, value)
	}

	// Get/save needed headers
	do.Language = do.Headers.Get(languageKey)
	do.tag = language.Make(do.Language)
	do.PluralForms = do.Headers.Get(pluralFormsKey)

	// Parse Plural-Forms formula
	if do.PluralForms == "" {
		return
	}

	// Split plural form header value
	pfs := strings.Split(do.PluralForms, ";")

	// Parse values
	for _, i := range pfs {
		vs := strings.SplitN(i, "=", 2)
		if len(vs) != 2 {
			continue
		}

		switch strings.TrimSpace(vs[0]) {
		case "nplurals":
			do.nplurals, _ = strconv.Atoi(vs[1])

		case "plural":
			do.plural = vs[1]

			if expr, err := plurals.Compile(do.plural); err == nil {
				do.pluralforms = expr
			}

		}
	}
}

// Drops any translations stored that have not been Set*() since 'po'
// was initialised
func (do *Domain) DropStaleTranslations() {
	do.trMutex.Lock()
	do.pluralMutex.Lock()
	defer do.trMutex.Unlock()
	defer do.pluralMutex.Unlock()

	for name, ctx := range do.contexts {
		for id, trans := range ctx {
			if trans.IsStale() {
				delete(ctx, id)
			}
		}
		if len(ctx) == 0 {
			delete(do.contexts, name)
		}
	}

	for id, trans := range do.translations {
		if trans.IsStale() {
			delete(do.translations, id)
		}
	}
}

// Set source references for a given translation
func (do *Domain) SetRefs(str string, refs []string) {
	do.trMutex.Lock()
	do.pluralMutex.Lock()
	defer do.trMutex.Unlock()
	defer do.pluralMutex.Unlock()

	if trans, ok := do.translations[str]; ok {
		trans.Refs = refs
	} else {
		trans = NewTranslation()
		trans.ID = str
		trans.SetRefs(refs)
		do.translations[str] = trans
	}
}

// Get source references for a given translation
func (do *Domain) GetRefs(str string) []string {
	// Sync read
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.translations != nil {
		if trans, ok := do.translations[str]; ok {
			return trans.Refs
		}
	}
	return nil
}

// Set the translation of a given string
func (do *Domain) Set(id, str string) {
	do.trMutex.Lock()
	do.pluralMutex.Lock()
	defer do.trMutex.Unlock()
	defer do.pluralMutex.Unlock()

	if trans, ok := do.translations[id]; ok {
		trans.Set(str)
	} else {
		trans = NewTranslation()
		trans.ID = id
		trans.Set(str)
		do.translations[str] = trans
	}
}

func (do *Domain) Get(str string, vars ...interface{}) string {
	// Sync read
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.translations != nil {
		if _, ok := do.translations[str]; ok {
			return Printf(do.translations[str].Get(), vars...)
		}
	}

	// Return the same we received by default
	return Printf(str, vars...)
}

// Set the (N)th plural form for the given string
func (do *Domain) SetN(id, plural string, n int, str string) {
	// Get plural form _before_ lock down
	pluralForm := do.pluralForm(n)

	do.trMutex.Lock()
	do.pluralMutex.Lock()
	defer do.trMutex.Unlock()
	defer do.pluralMutex.Unlock()

	if trans, ok := do.translations[id]; ok {
		trans.SetN(pluralForm, str)
	} else {
		trans = NewTranslation()
		trans.ID = id
		trans.PluralID = plural
		trans.SetN(pluralForm, str)
		do.translations[str] = trans
	}
}

// GetN retrieves the (N)th plural form of Translation for the given string.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (do *Domain) GetN(str, plural string, n int, vars ...interface{}) string {
	// Sync read
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.translations != nil {
		if _, ok := do.translations[str]; ok {
			return Printf(do.translations[str].GetN(do.pluralForm(n)), vars...)
		}
	}

	// Parse plural forms to distinguish between plural and singular
	if do.pluralForm(n) == 0 {
		return Printf(str, vars...)
	}
	return Printf(plural, vars...)
}

// Set the translation for the given string in the given context
func (do *Domain) SetC(id, ctx, str string) {
	do.trMutex.Lock()
	do.pluralMutex.Lock()
	defer do.trMutex.Unlock()
	defer do.pluralMutex.Unlock()

	if context, ok := do.contexts[ctx]; ok {
		if trans, hasTrans := context[id]; hasTrans {
			trans.Set(str)
		} else {
			trans = NewTranslation()
			trans.ID = id
			trans.Set(str)
			context[id] = trans
		}
	} else {
		trans := NewTranslation()
		trans.ID = id
		trans.Set(str)
		do.contexts[ctx] = map[string]*Translation{
			id: trans,
		}
	}
}

// GetC retrieves the corresponding Translation for a given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (do *Domain) GetC(str, ctx string, vars ...interface{}) string {
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.contexts != nil {
		if _, ok := do.contexts[ctx]; ok {
			if do.contexts[ctx] != nil {
				if _, ok := do.contexts[ctx][str]; ok {
					return Printf(do.contexts[ctx][str].Get(), vars...)
				}
			}
		}
	}

	// Return the string we received by default
	return Printf(str, vars...)
}

// Set the (N)th plural form for the given string in the given context
func (do *Domain) SetNC(id, plural, ctx string, n int, str string) {
	// Get plural form _before_ lock down
	pluralForm := do.pluralForm(n)

	do.trMutex.Lock()
	do.pluralMutex.Lock()
	defer do.trMutex.Unlock()
	defer do.pluralMutex.Unlock()

	if context, ok := do.contexts[ctx]; ok {
		if trans, hasTrans := context[id]; hasTrans {
			trans.SetN(pluralForm, str)
		} else {
			trans = NewTranslation()
			trans.ID = id
			trans.SetN(pluralForm, str)
			context[id] = trans
		}
	} else {
		trans := NewTranslation()
		trans.ID = id
		trans.SetN(pluralForm, str)
		do.contexts[ctx] = map[string]*Translation{
			id: trans,
		}
	}
}

// GetNC retrieves the (N)th plural form of Translation for the given string in the given context.
// Supports optional parameters (vars... interface{}) to be inserted on the formatted string using the fmt.Printf syntax.
func (do *Domain) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	if do.contexts != nil {
		if _, ok := do.contexts[ctx]; ok {
			if do.contexts[ctx] != nil {
				if _, ok := do.contexts[ctx][str]; ok {
					return Printf(do.contexts[ctx][str].GetN(do.pluralForm(n)), vars...)
				}
			}
		}
	}

	if n == 1 {
		return Printf(str, vars...)
	}
	return Printf(plural, vars...)
}

//GetTranslations returns a copy of every translation in the domain. It does not support contexts.
func (do *Domain) GetTranslations() map[string]*Translation {
	all := make(map[string]*Translation, len(do.translations))

	do.trMutex.RLock()
	defer do.trMutex.RUnlock()

	for msgID, trans := range do.translations {
		newTrans := NewTranslation()
		newTrans.ID = trans.ID
		newTrans.PluralID = trans.PluralID
		newTrans.dirty = trans.dirty
		if len(trans.Refs) > 0 {
			newTrans.Refs = make([]string, len(trans.Refs))
			copy(newTrans.Refs, trans.Refs)
		}
		for k, v := range trans.Trs {
			newTrans.Trs[k] = v
		}
		all[msgID] = newTrans
	}

	return all
}

type SourceReference struct {
	path    string
	line    int
	context string
	trans   *Translation
}

func extractPathAndLine(ref string) (string, int) {
	var path string
	var line int
	colonIdx := strings.IndexRune(ref, ':')
	if colonIdx >= 0 {
		path = ref[:colonIdx]
		line, _ = strconv.Atoi(ref[colonIdx+1:])
	} else {
		path = ref
		line = 0
	}
	return path, line
}

// MarshalText implements encoding.TextMarshaler interface
// Assists round-trip of POT/PO content
func (do *Domain) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	if len(do.headerComments) > 0 {
		buf.WriteString(strings.Join(do.headerComments, "\n"))
		buf.WriteByte(byte('\n'))
	}
	buf.WriteString("msgid \"\"\nmsgstr \"\"")

	// Standard order consistent with xgettext
	headerOrder := map[string]int{
		"project-id-version":        0,
		"report-msgid-bugs-to":      1,
		"pot-creation-date":         2,
		"po-revision-date":          3,
		"last-translator":           4,
		"language-team":             5,
		"language":                  6,
		"mime-version":              7,
		"content-type":              9,
		"content-transfer-encoding": 10,
		"plural-forms":              11,
	}

	headerKeys := make([]string, 0, len(do.Headers))

	for k, _ := range do.Headers {
		headerKeys = append(headerKeys, k)
	}

	sort.Slice(headerKeys, func(i, j int) bool {
		var iOrder int
		var jOrder int
		var ok bool
		if iOrder, ok = headerOrder[strings.ToLower(headerKeys[i])]; !ok {
			iOrder = 8
		}

		if jOrder, ok = headerOrder[strings.ToLower(headerKeys[j])]; !ok {
			jOrder = 8
		}

		if iOrder < jOrder {
			return true
		}
		if iOrder > jOrder {
			return false
		}
		return headerKeys[i] < headerKeys[j]
	})

	for _, k := range headerKeys {
		// Access Headers map directly so as not to canonicalise
		v := do.Headers[k]

		for _, value := range v {
			buf.WriteString("\n\"" + k + ": " + value + "\\n\"")
		}
	}

	// Just as with headers, output translations in consistent order (to minimise diffs between round-trips), with (first) source reference taking priority, followed by context and finally ID
	references := make([]SourceReference, 0)
	for name, ctx := range do.contexts {
		for id, trans := range ctx {
			if id == "" {
				continue
			}
			if len(trans.Refs) > 0 {
				path, line := extractPathAndLine(trans.Refs[0])
				references = append(references, SourceReference{
					path,
					line,
					name,
					trans,
				})
			} else {
				references = append(references, SourceReference{
					"",
					0,
					name,
					trans,
				})
			}
		}
	}

	for id, trans := range do.translations {
		if id == "" {
			continue
		}

		if len(trans.Refs) > 0 {
			path, line := extractPathAndLine(trans.Refs[0])
			references = append(references, SourceReference{
				path,
				line,
				"",
				trans,
			})
		} else {
			references = append(references, SourceReference{
				"",
				0,
				"",
				trans,
			})
		}
	}

	sort.Slice(references, func(i, j int) bool {
		if references[i].path < references[j].path {
			return true
		}
		if references[i].path > references[j].path {
			return false
		}
		if references[i].line < references[j].line {
			return true
		}
		if references[i].line > references[j].line {
			return false
		}

		if references[i].context < references[j].context {
			return true
		}
		if references[i].context > references[j].context {
			return false
		}
		return references[i].trans.ID < references[j].trans.ID
	})

	for _, ref := range references {
		trans := ref.trans
		if len(trans.Refs) > 0 {
			buf.WriteString("\n\n#: " + strings.Join(trans.Refs, " "))
		} else {
			buf.WriteByte(byte('\n'))
		}

		if ref.context == "" {
			buf.WriteString("\nmsgid \"" + trans.ID + "\"")
		} else {
			buf.WriteString("\nmsgctxt \"" + ref.context + "\"\nmsgid \"" + trans.ID + "\"")
		}

		if trans.PluralID == "" {
			buf.WriteString("\nmsgstr \"" + trans.Trs[0] + "\"")
		} else {
			buf.WriteString("\nmsgid_plural \"" + trans.PluralID + "\"")
			for i, tr := range trans.Trs {
				buf.WriteString("\nmsgstr[" + strconv.Itoa(i) + "] \"" + tr + "\"")
			}
		}
	}

	return buf.Bytes(), nil
}

// MarshalBinary implements encoding.BinaryMarshaler interface
func (do *Domain) MarshalBinary() ([]byte, error) {
	obj := new(TranslatorEncoding)
	obj.Headers = do.Headers
	obj.Language = do.Language
	obj.PluralForms = do.PluralForms
	obj.Nplurals = do.nplurals
	obj.Plural = do.plural
	obj.Translations = do.translations
	obj.Contexts = do.contexts

	var buff bytes.Buffer
	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(obj)

	return buff.Bytes(), err
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler interface
func (do *Domain) UnmarshalBinary(data []byte) error {
	buff := bytes.NewBuffer(data)
	obj := new(TranslatorEncoding)

	decoder := gob.NewDecoder(buff)
	err := decoder.Decode(obj)
	if err != nil {
		return err
	}

	do.Headers = obj.Headers
	do.Language = obj.Language
	do.PluralForms = obj.PluralForms
	do.nplurals = obj.Nplurals
	do.plural = obj.Plural
	do.translations = obj.Translations
	do.contexts = obj.Contexts

	if expr, err := plurals.Compile(do.plural); err == nil {
		do.pluralforms = expr
	}

	return nil
}
