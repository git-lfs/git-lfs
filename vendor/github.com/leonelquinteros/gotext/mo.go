/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"bytes"
	"encoding/binary"
)

const (
	// MoMagicLittleEndian encoding
	MoMagicLittleEndian = 0x950412de
	// MoMagicBigEndian encoding
	MoMagicBigEndian = 0xde120495

	// EotSeparator msgctxt and msgid separator
	EotSeparator = "\x04"
	// NulSeparator msgid and msgstr separator
	NulSeparator = "\x00"
)

/*
Mo parses the content of any MO file and provides all the Translation functions needed.
It's the base object used by all package methods.
And it's safe for concurrent use by multiple goroutines by using the sync package for locking.

Example:

	import (
		"fmt"
		"github.com/leonelquinteros/gotext"
	)

	func main() {
		// Create po object
		po := gotext.NewMoTranslator()

		// Parse .po file
		po.ParseFile("/path/to/po/file/translations.mo")

		// Get Translation
		fmt.Println(po.Get("Translate this"))
	}

*/
type Mo struct {
	//these three public members are for backwards compatibility. they are just set to the value in the domain
	Headers     HeaderMap
	Language    string
	PluralForms string
	domain      *Domain
}

//NewMo should always be used to instantiate a new Mo object
func NewMo() *Mo {
	mo := new(Mo)
	mo.domain = NewDomain()

	return mo
}

func (mo *Mo) GetDomain() *Domain {
	return mo.domain
}

//all of the Get functions are for convenience and aid in backwards compatibility
func (mo *Mo) Get(str string, vars ...interface{}) string {
	return mo.domain.Get(str, vars...)
}

func (mo *Mo) GetN(str, plural string, n int, vars ...interface{}) string {
	return mo.domain.GetN(str, plural, n, vars...)
}

func (mo *Mo) GetC(str, ctx string, vars ...interface{}) string {
	return mo.domain.GetC(str, ctx, vars...)
}

func (mo *Mo) GetNC(str, plural string, n int, ctx string, vars ...interface{}) string {
	return mo.domain.GetNC(str, plural, n, ctx, vars...)
}

func (mo *Mo) MarshalBinary() ([]byte, error) {
	return mo.domain.MarshalBinary()
}

func (mo *Mo) UnmarshalBinary(data []byte) error {
	return mo.domain.UnmarshalBinary(data)
}

func (mo *Mo) ParseFile(f string) {
	data, err := getFileData(f)
	if err != nil {
		return
	}

	mo.Parse(data)
}

// Parse loads the translations specified in the provided byte slice, in the GNU gettext .mo format
func (mo *Mo) Parse(buf []byte) {
	// Lock while parsing
	mo.domain.trMutex.Lock()
	mo.domain.pluralMutex.Lock()
	defer mo.domain.trMutex.Unlock()
	defer mo.domain.pluralMutex.Unlock()

	r := bytes.NewReader(buf)

	var magicNumber uint32
	if err := binary.Read(r, binary.LittleEndian, &magicNumber); err != nil {
		return
		// return fmt.Errorf("gettext: %v", err)
	}
	var bo binary.ByteOrder
	switch magicNumber {
	case MoMagicLittleEndian:
		bo = binary.LittleEndian
	case MoMagicBigEndian:
		bo = binary.BigEndian
	default:
		return
		// return fmt.Errorf("gettext: %v", "invalid magic number")
	}

	var header struct {
		MajorVersion uint16
		MinorVersion uint16
		MsgIDCount   uint32
		MsgIDOffset  uint32
		MsgStrOffset uint32
		HashSize     uint32
		HashOffset   uint32
	}
	if err := binary.Read(r, bo, &header); err != nil {
		return
		// return fmt.Errorf("gettext: %v", err)
	}
	if v := header.MajorVersion; v != 0 && v != 1 {
		return
		// return fmt.Errorf("gettext: %v", "invalid version number")
	}
	if v := header.MinorVersion; v != 0 && v != 1 {
		return
		// return fmt.Errorf("gettext: %v", "invalid version number")
	}

	msgIDStart := make([]uint32, header.MsgIDCount)
	msgIDLen := make([]uint32, header.MsgIDCount)
	if _, err := r.Seek(int64(header.MsgIDOffset), 0); err != nil {
		return
		// return fmt.Errorf("gettext: %v", err)
	}
	for i := 0; i < int(header.MsgIDCount); i++ {
		if err := binary.Read(r, bo, &msgIDLen[i]); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}
		if err := binary.Read(r, bo, &msgIDStart[i]); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}
	}

	msgStrStart := make([]int32, header.MsgIDCount)
	msgStrLen := make([]int32, header.MsgIDCount)
	if _, err := r.Seek(int64(header.MsgStrOffset), 0); err != nil {
		return
		// return fmt.Errorf("gettext: %v", err)
	}
	for i := 0; i < int(header.MsgIDCount); i++ {
		if err := binary.Read(r, bo, &msgStrLen[i]); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}
		if err := binary.Read(r, bo, &msgStrStart[i]); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}
	}

	for i := 0; i < int(header.MsgIDCount); i++ {
		if _, err := r.Seek(int64(msgIDStart[i]), 0); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}
		msgIDData := make([]byte, msgIDLen[i])
		if _, err := r.Read(msgIDData); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}

		if _, err := r.Seek(int64(msgStrStart[i]), 0); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}
		msgStrData := make([]byte, msgStrLen[i])
		if _, err := r.Read(msgStrData); err != nil {
			return
			// return fmt.Errorf("gettext: %v", err)
		}

		if len(msgIDData) == 0 {
			mo.addTranslation(msgIDData, msgStrData)
		} else {
			mo.addTranslation(msgIDData, msgStrData)
		}
	}

	// Parse headers
	mo.domain.parseHeaders()

	// set values on this struct
	// this is for backwards compatibility
	mo.Language = mo.domain.Language
	mo.PluralForms = mo.domain.PluralForms
	mo.Headers = mo.domain.Headers
}

func (mo *Mo) addTranslation(msgid, msgstr []byte) {
	translation := NewTranslation()
	var msgctxt []byte
	var msgidPlural []byte

	d := bytes.Split(msgid, []byte(EotSeparator))
	if len(d) == 1 {
		msgid = d[0]
	} else {
		msgid, msgctxt = d[1], d[0]
	}

	dd := bytes.Split(msgid, []byte(NulSeparator))
	if len(dd) > 1 {
		msgid = dd[0]
		dd = dd[1:]
	}

	translation.ID = string(msgid)

	msgidPlural = bytes.Join(dd, []byte(NulSeparator))
	if len(msgidPlural) > 0 {
		translation.PluralID = string(msgidPlural)
	}

	ddd := bytes.Split(msgstr, []byte(NulSeparator))
	if len(ddd) > 0 {
		for i, s := range ddd {
			translation.Trs[i] = string(s)
		}
	}

	if len(msgctxt) > 0 {
		// With context...
		if _, ok := mo.domain.contexts[string(msgctxt)]; !ok {
			mo.domain.contexts[string(msgctxt)] = make(map[string]*Translation)
		}
		mo.domain.contexts[string(msgctxt)][translation.ID] = translation
	} else {
		mo.domain.translations[translation.ID] = translation
	}
}
