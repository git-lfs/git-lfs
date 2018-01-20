package humanize

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/git-lfs/git-lfs/errors"
)

const (
	Byte = 1 << (iota * 10)
	Kibibyte
	Mebibyte
	Gibibyte
	Tebibyte
	Pebibyte

	Kilobyte = 1000 * Byte
	Megabyte = 1000 * Kilobyte
	Gigabyte = 1000 * Megabyte
	Terabyte = 1000 * Gigabyte
	Petabyte = 1000 * Terabyte

	// eps is the machine epsilon, or a 64-bit floating point value
	// reasonably close to zero.
	eps float64 = 7.0/3 - 4.0/3 - 1.0
)

var bytesTable = map[string]uint64{
	"":  Byte,
	"b": Byte,

	"kib": Kibibyte,
	"mib": Mebibyte,
	"gib": Gibibyte,
	"tib": Tebibyte,
	"pib": Pebibyte,

	"kb": Kilobyte,
	"mb": Megabyte,
	"gb": Gigabyte,
	"tb": Terabyte,
	"pb": Petabyte,
}

// ParseBytes parses a given human-readable bytes or ibytes string into a number
// of bytes, or an error if the string was unable to be parsed.
func ParseBytes(str string) (uint64, error) {
	var sep int
	for _, r := range str {
		if !(unicode.IsDigit(r) || r == '.' || r == ',') {
			break
		}

		sep = sep + 1
	}

	var f float64

	if s := strings.Replace(str[:sep], ",", "", -1); len(s) > 0 {
		var err error

		f, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
	}

	m, err := ParseByteUnit(str[sep:])
	if err != nil {
		return 0, err
	}

	f = f * float64(m)
	if f >= math.MaxUint64 {
		return 0, errors.New("number of bytes too large")
	}
	return uint64(f), nil
}

// ParseByteUnit returns the number of bytes in a given unit of storage, or an
// error, if that unit is unrecognized.
func ParseByteUnit(str string) (uint64, error) {
	str = strings.TrimSpace(str)
	str = strings.ToLower(str)

	if u, ok := bytesTable[str]; ok {
		return u, nil
	}
	return 0, errors.Errorf("unknown unit: %q", str)
}

var sizes = []string{"B", "KB", "MB", "GB", "TB", "PB"}

// FormatBytes outputs the given number of bytes "s" as a human-readable string,
// rounding to the nearest half within .01.
func FormatBytes(s uint64) string {
	var e float64
	if s == 0 {
		e = 0
	} else {
		e = math.Floor(log(float64(s), 1000))
	}

	unit := uint64(math.Pow(1000, e))
	suffix := sizes[int(e)]

	return fmt.Sprintf("%s %s",
		FormatBytesUnit(s, unit), suffix)
}

// FormatBytesUnit outputs the given number of bytes "s" as a quantity of the
// given units "u" to the nearest half within .01.
func FormatBytesUnit(s, u uint64) string {
	var rounded float64
	if s < 10 {
		rounded = float64(s)
	} else {
		rounded = math.Floor(float64(s)/float64(u)*10+.5) / 10
	}

	format := "%.0f"
	if rounded < 10 && u > 1 {
		format = "%.1f"
	}

	return fmt.Sprintf(format, rounded)
}

// FormatByteRate outputs the given rate of transfer "r" as the quotient of "s"
// (the number of bytes transferred) over "d" (the duration of time that those
// bytes were transferred in).
//
// It displays the output as a quantity of a "per-unit-time" unit (i.e., B/s,
// MiB/s) in the most representative fashion possible, as above.
func FormatByteRate(s uint64, d time.Duration) string {
	// e is the index of the most representative unit of storage.
	var e float64

	// f is the floating-point equivalent of "s", so as to avoid more
	// conversions than necessary.
	f := float64(s)

	if f != 0 {
		f = f / d.Seconds()
		e = math.Floor(log(f, 1000))
		if e <= eps {
			// The result of math.Floor(log(r, 1000)) can be
			// "close-enough" to zero that it should be effectively
			// considered zero.
			e = 0
		}
	}

	unit := uint64(math.Pow(1000, e))
	suffix := sizes[int(e)]

	return fmt.Sprintf("%s %s/s",
		FormatBytesUnit(uint64(math.Ceil(f)), unit), suffix)
}

// log takes the log base "b" of "n" (\log_b{n})
func log(n, b float64) float64 {
	return math.Log(n) / math.Log(b)
}
