/*
 * Copyright (c) 2018 DeineAgentur UG https://www.deineagentur.com. All rights reserved.
 * Licensed under the MIT License. See LICENSE file in the project root for full license information.
 */

package gotext

import (
	"errors"
	"io/ioutil"
	"os"
)

// Translator interface is used by Locale and Po objects.Translator
// It contains all methods needed to parse translation sources and obtain corresponding translations.
// Also implements gob.GobEncoder/gob.DobDecoder interfaces to allow serialization of Locale objects.
type Translator interface {
	ParseFile(f string)
	Parse(buf []byte)
	Get(str string, vars ...interface{}) string
	GetN(str, plural string, n int, vars ...interface{}) string
	GetC(str, ctx string, vars ...interface{}) string
	GetNC(str, plural string, n int, ctx string, vars ...interface{}) string

	MarshalBinary() ([]byte, error)
	UnmarshalBinary([]byte) error
	GetDomain() *Domain
}

// TranslatorEncoding is used as intermediary storage to encode Translator objects to Gob.
type TranslatorEncoding struct {
	// Headers storage
	Headers HeaderMap

	// Language header
	Language string

	// Plural-Forms header
	PluralForms string

	// Parsed Plural-Forms header values
	Nplurals int
	Plural   string

	// Storage
	Translations map[string]*Translation
	Contexts     map[string]map[string]*Translation
}

// GetTranslator is used to recover a Translator object after unmarshalling the TranslatorEncoding object.
// Internally uses a Po object as it should be switchable with Mo objects without problem.
// External Translator implementations should be able to serialize into a TranslatorEncoding object in order to
// deserialize into a Po-compatible object.
func (te *TranslatorEncoding) GetTranslator() Translator {
	po := NewPo()
	po.domain = NewDomain()
	po.domain.Headers = te.Headers
	po.domain.Language = te.Language
	po.domain.PluralForms = te.PluralForms
	po.domain.nplurals = te.Nplurals
	po.domain.plural = te.Plural
	po.domain.translations = te.Translations
	po.domain.contexts = te.Contexts

	return po
}

//getFileData reads a file and returns the byte slice after doing some basic sanity checking
func getFileData(f string) ([]byte, error) {
	// Check if file exists
	info, err := os.Stat(f)
	if err != nil {
		return nil, err
	}

	// Check that isn't a directory
	if info.IsDir() {
		return nil, errors.New("cannot parse a directory")
	}

	return ioutil.ReadFile(f)
}
