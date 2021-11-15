[![GitHub release](https://img.shields.io/github/release/leonelquinteros/gotext.svg)](https://github.com/leonelquinteros/gotext)
[![MIT license](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
![Gotext build](https://github.com/leonelquinteros/gotext/workflows/Gotext%20build/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/leonelquinteros/gotext)](https://goreportcard.com/report/github.com/leonelquinteros/gotext)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/leonelquinteros/gotext)](https://pkg.go.dev/github.com/leonelquinteros/gotext)


# Gotext

[GNU gettext utilities](https://www.gnu.org/software/gettext) for Go.


# Features

- Implements GNU gettext support in native Go.
- Complete support for [PO files](https://www.gnu.org/software/gettext/manual/html_node/PO-Files.html) including:
  - Support for multiline strings and headers.
  - Support for variables inside translation strings using Go's [fmt syntax](https://golang.org/pkg/fmt/).
  - Support for [pluralization rules](https://www.gnu.org/software/gettext/manual/html_node/Translating-plural-forms.html).
  - Support for [message contexts](https://www.gnu.org/software/gettext/manual/html_node/Contexts.html).
- Support for MO files.
- Thread-safe: This package is safe for concurrent use across multiple goroutines.
- It works with UTF-8 encoding as it's the default for Go language.
- Unit tests available.
- Language codes are automatically simplified from the form `en_UK` to `en` if the first isn't available.
- Ready to use inside Go templates.
- Objects are serializable to []byte to store them in cache.
- Support for Go Modules.


# License

[MIT license](LICENSE)


# Documentation

Refer to the Godoc package documentation at (https://godoc.org/github.com/leonelquinteros/gotext)


# Installation

```
go get github.com/leonelquinteros/gotext
```

- There are no requirements or dependencies to use this package.
- No need to install GNU gettext utilities (unless specific needs of CLI tools).
- No need for environment variables. Some naming conventions are applied but not needed.


## Version vendoring

Stable releases use [semantic versioning](http://semver.org/spec/v2.0.0.html) tagging on this repository.

You can rely on this to use your preferred vendoring tool or to manually retrieve the corresponding release tag from the GitHub repository.


### Vendoring with [Go Modules](https://github.com/golang/go/wiki/Modules) (Recommended)

Add `github.com/leonelquinteros/gotext` inside the  `require` section in your `go.mod` file.

i.e.
```
require (
    github.com/leonelquinteros/gotext v1.4.0
)
```


### Vendoring with [dep](https://golang.github.io/dep/)

To use last stable version (v1.4.0 at the moment of writing)

```
dep ensure -add github.com/leonelquinteros/gotext@v1.4.0
```

Import as

```go
import "github.com/leonelquinteros/gotext"
```


### Vendoring with [gopkg.in](http://labix.org/gopkg.in)

[http://gopkg.in/leonelquinteros/gotext.v1](http://gopkg.in/leonelquinteros/gotext.v1)

To get the latest v1 package stable release, execute:

```
go get gopkg.in/leonelquinteros/gotext.v1
```

Import as

```go
import "gopkg.in/leonelquinteros/gotext.v1"
```

Refer to it as gotext.


# Locales directories structure

The package will assume a directories structure starting with a base path that will be provided to the package configuration
or to object constructors depending on the use, but either will use the same convention to lookup inside the base path.

Inside the base directory where will be the language directories named using the language and country 2-letter codes (en_US, es_AR, ...).
All package functions can lookup after the simplified version for each language in case the full code isn't present but the more general language code exists.
So if the language set is `en_UK`, but there is no directory named after that code and there is a directory named `en`,
all package functions will be able to resolve this generalization and provide translations for the more general library.

The language codes are assumed to be [ISO 639-1](https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes) codes (2-letter codes).
That said, most functions will work with any coding standard as long the directory name matches the language code set on the configuration.

Then, there can be a `LC_MESSAGES` containing all PO files or the PO files themselves.
A library directory structure can look like:

```
/path/to/locales
/path/to/locales/en_US
/path/to/locales/en_US/LC_MESSAGES
/path/to/locales/en_US/LC_MESSAGES/default.po
/path/to/locales/en_US/LC_MESSAGES/extras.po
/path/to/locales/en_UK
/path/to/locales/en_UK/LC_MESSAGES
/path/to/locales/en_UK/LC_MESSAGES/default.po
/path/to/locales/en_UK/LC_MESSAGES/extras.po
/path/to/locales/en_AU
/path/to/locales/en_AU/LC_MESSAGES
/path/to/locales/en_AU/LC_MESSAGES/default.po
/path/to/locales/en_AU/LC_MESSAGES/extras.po
/path/to/locales/es
/path/to/locales/es/default.po
/path/to/locales/es/extras.po
/path/to/locales/es_ES
/path/to/locales/es_ES/default.po
/path/to/locales/es_ES/extras.po
/path/to/locales/fr
/path/to/locales/fr/default.po
/path/to/locales/fr/extras.po
```

And so on...


# Usage examples

## Using package for single language/domain settings

For quick/simple translations you can use the package level functions directly.

```go
import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // Configure package
    gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

    // Translate text from default domain
    fmt.Println(gotext.Get("My text on 'domain-name' domain"))

    // Translate text from a different domain without reconfigure
    fmt.Println(gotext.GetD("domain2", "Another text on a different domain"))
}

```

## Using dynamic variables on translations

All translation strings support dynamic variables to be inserted without translate.
Use the fmt.Printf syntax (from Go's "fmt" package) to specify how to print the non-translated variable inside the translation string.

```go
import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // Configure package
    gotext.Configure("/path/to/locales/root/dir", "en_UK", "domain-name")

    // Set variables
    name := "John"

    // Translate text with variables
    fmt.Println(gotext.Get("Hi, my name is %s", name))
}

```


## Using Locale object

When having multiple languages/domains/libraries at the same time, you can create Locale objects for each variation
so you can handle each settings on their own.

```go
import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // Create Locale with library path and language code
    l := gotext.NewLocale("/path/to/locales/root/dir", "es_UY")

    // Load domain '/path/to/locales/root/dir/es_UY/default.po'
    l.AddDomain("default")

    // Translate text from default domain
    fmt.Println(l.Get("Translate this"))

    // Load different domain
    l.AddDomain("translations")

    // Translate text from domain
    fmt.Println(l.GetD("translations", "Translate this"))
}
```

This is also helpful for using inside templates (from the "text/template" package), where you can pass the Locale object to the template.
If you set the Locale object as "Loc" in the template, then the template code would look like:

```
{{ .Loc.Get "Translate this" }}
```


## Using the Po object to handle .po files and PO-formatted strings

For when you need to work with PO files and strings,
you can directly use the Po object to parse it and access the translations in there in the same way.

```go
import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // Set PO content
    str := `
msgid "Translate this"
msgstr "Translated text"

msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgstr "This one sets the var: %s"
`

    // Create Po object
    po := new(gotext.Po)
    po.Parse(str)

    fmt.Println(po.Get("Translate this"))
}
```


## Use plural forms of translations

PO format supports defining one or more plural forms for the same translation.
Relying on the PO file headers, a Plural-Forms formula can be set on the translation file
as defined in (https://www.gnu.org/savannah-checkouts/gnu/gettext/manual/html_node/Plural-forms.html)

```go
import (
    "fmt"
    "github.com/leonelquinteros/gotext"
)

func main() {
    // Set PO content
    str := `
msgid ""
msgstr ""

# Header below
"Plural-Forms: nplurals=2; plural=(n != 1);\n"

msgid "Translate this"
msgstr "Translated text"

msgid "Another string"
msgstr ""

msgid "One with var: %s"
msgid_plural "Several with vars: %s"
msgstr[0] "This one is the singular: %s"
msgstr[1] "This one is the plural: %s"
`

    // Create Po object
    po := new(gotext.Po)
    po.Parse(str)

    fmt.Println(po.GetN("One with var: %s", "Several with vars: %s", 54, v))
    // "This one is the plural: Variable"
}
```


# Contribute

- Please, contribute.
- Use the package on your projects.
- Report issues on Github.
- Send pull requests for bugfixes and improvements.
- Send proposals on Github issues.

