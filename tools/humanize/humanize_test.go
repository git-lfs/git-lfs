package humanize_test

import (
	"math"
	"testing"

	"github.com/git-lfs/git-lfs/tools/humanize"
	"github.com/stretchr/testify/assert"
)

type ParseBytesTestCase struct {
	Given    string
	Expected uint64
	Err      error
}

func (c *ParseBytesTestCase) Assert(t *testing.T) {
	got, err := humanize.ParseBytes(c.Given)
	if c.Err == nil {
		assert.NoError(t, err, "unexpected error: %s", err)
		assert.EqualValues(t, c.Expected, got)
	} else {
		assert.Equal(t, c.Err, err)
	}
}

type FormatBytesTestCase struct {
	Given    uint64
	Expected string
}

func (c *FormatBytesTestCase) Assert(t *testing.T) {
	assert.Equal(t, c.Expected, humanize.FormatBytes(c.Given))
}

type ParseByteUnitTestCase struct {
	Given    string
	Expected uint64
	Err      string
}

func (c *ParseByteUnitTestCase) Assert(t *testing.T) {
	got, err := humanize.ParseByteUnit(c.Given)
	if len(c.Err) == 0 {
		assert.NoError(t, err, "unexpected error: %s", err)
		assert.EqualValues(t, c.Expected, got)
	} else {
		assert.EqualError(t, err, c.Err)
	}
}

type FormatBytesUnitTestCase struct {
	Given    uint64
	Unit     uint64
	Expected string
}

func (c *FormatBytesUnitTestCase) Assert(t *testing.T) {
	assert.Equal(t, c.Expected, humanize.FormatBytesUnit(c.Given, c.Unit))
}

func TestParseBytes(t *testing.T) {
	for desc, c := range map[string]*ParseBytesTestCase{
		"parse byte":     {"10B", uint64(10 * math.Pow(2, 0)), nil},
		"parse kibibyte": {"20KIB", uint64(20 * math.Pow(2, 10)), nil},
		"parse mebibyte": {"30MIB", uint64(30 * math.Pow(2, 20)), nil},
		"parse gibibyte": {"40GIB", uint64(40 * math.Pow(2, 30)), nil},
		"parse tebibyte": {"50TIB", uint64(50 * math.Pow(2, 40)), nil},
		"parse pebibyte": {"60PIB", uint64(60 * math.Pow(2, 50)), nil},

		"parse byte (lowercase)":     {"10b", uint64(10 * math.Pow(2, 0)), nil},
		"parse kibibyte (lowercase)": {"20kib", uint64(20 * math.Pow(2, 10)), nil},
		"parse mebibyte (lowercase)": {"30mib", uint64(30 * math.Pow(2, 20)), nil},
		"parse gibibyte (lowercase)": {"40gib", uint64(40 * math.Pow(2, 30)), nil},
		"parse tebibyte (lowercase)": {"50tib", uint64(50 * math.Pow(2, 40)), nil},
		"parse pebibyte (lowercase)": {"60pib", uint64(60 * math.Pow(2, 50)), nil},

		"parse byte (with space)":     {"10 B", uint64(10 * math.Pow(2, 0)), nil},
		"parse kibibyte (with space)": {"20 KIB", uint64(20 * math.Pow(2, 10)), nil},
		"parse mebibyte (with space)": {"30 MIB", uint64(30 * math.Pow(2, 20)), nil},
		"parse gibibyte (with space)": {"40 GIB", uint64(40 * math.Pow(2, 30)), nil},
		"parse tebibyte (with space)": {"50 TIB", uint64(50 * math.Pow(2, 40)), nil},
		"parse pebibyte (with space)": {"60 PIB", uint64(60 * math.Pow(2, 50)), nil},

		"parse byte (with space, lowercase)":     {"10 b", uint64(10 * math.Pow(2, 0)), nil},
		"parse kibibyte (with space, lowercase)": {"20 kib", uint64(20 * math.Pow(2, 10)), nil},
		"parse mebibyte (with space, lowercase)": {"30 mib", uint64(30 * math.Pow(2, 20)), nil},
		"parse gibibyte (with space, lowercase)": {"40 gib", uint64(40 * math.Pow(2, 30)), nil},
		"parse tebibyte (with space, lowercase)": {"50 tib", uint64(50 * math.Pow(2, 40)), nil},
		"parse pebibyte (with space, lowercase)": {"60 pib", uint64(60 * math.Pow(2, 50)), nil},

		"parse kilobyte": {"20KB", uint64(20 * math.Pow(10, 3)), nil},
		"parse megabyte": {"30MB", uint64(30 * math.Pow(10, 6)), nil},
		"parse gigabyte": {"40GB", uint64(40 * math.Pow(10, 9)), nil},
		"parse terabyte": {"50TB", uint64(50 * math.Pow(10, 12)), nil},
		"parse petabyte": {"60PB", uint64(60 * math.Pow(10, 15)), nil},

		"parse kilobyte (lowercase)": {"20kb", uint64(20 * math.Pow(10, 3)), nil},
		"parse megabyte (lowercase)": {"30mb", uint64(30 * math.Pow(10, 6)), nil},
		"parse gigabyte (lowercase)": {"40gb", uint64(40 * math.Pow(10, 9)), nil},
		"parse terabyte (lowercase)": {"50tb", uint64(50 * math.Pow(10, 12)), nil},
		"parse petabyte (lowercase)": {"60pb", uint64(60 * math.Pow(10, 15)), nil},

		"parse kilobyte (with space)": {"20 KB", uint64(20 * math.Pow(10, 3)), nil},
		"parse megabyte (with space)": {"30 MB", uint64(30 * math.Pow(10, 6)), nil},
		"parse gigabyte (with space)": {"40 GB", uint64(40 * math.Pow(10, 9)), nil},
		"parse terabyte (with space)": {"50 TB", uint64(50 * math.Pow(10, 12)), nil},
		"parse petabyte (with space)": {"60 PB", uint64(60 * math.Pow(10, 15)), nil},

		"parse kilobyte (with space, lowercase)": {"20 kb", uint64(20 * math.Pow(10, 3)), nil},
		"parse megabyte (with space, lowercase)": {"30 mb", uint64(30 * math.Pow(10, 6)), nil},
		"parse gigabyte (with space, lowercase)": {"40 gb", uint64(40 * math.Pow(10, 9)), nil},
		"parse terabyte (with space, lowercase)": {"50 tb", uint64(50 * math.Pow(10, 12)), nil},
		"parse petabyte (with space, lowercase)": {"60 pb", uint64(60 * math.Pow(10, 15)), nil},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestFormatBytes(t *testing.T) {
	for desc, c := range map[string]*FormatBytesTestCase{
		"format bytes":     {uint64(1 * math.Pow(10, 0)), "1 B"},
		"format kilobytes": {uint64(1 * math.Pow(10, 3)), "1.0 KB"},
		"format megabytes": {uint64(1 * math.Pow(10, 6)), "1.0 MB"},
		"format gigabytes": {uint64(1 * math.Pow(10, 9)), "1.0 GB"},
		"format petabytes": {uint64(1 * math.Pow(10, 12)), "1.0 TB"},
		"format terabytes": {uint64(1 * math.Pow(10, 15)), "1.0 PB"},

		"format kilobytes under": {uint64(1.49 * math.Pow(10, 3)), "1.5 KB"},
		"format megabytes under": {uint64(1.49 * math.Pow(10, 6)), "1.5 MB"},
		"format gigabytes under": {uint64(1.49 * math.Pow(10, 9)), "1.5 GB"},
		"format petabytes under": {uint64(1.49 * math.Pow(10, 12)), "1.5 TB"},
		"format terabytes under": {uint64(1.49 * math.Pow(10, 15)), "1.5 PB"},

		"format kilobytes over": {uint64(1.51 * math.Pow(10, 3)), "1.5 KB"},
		"format megabytes over": {uint64(1.51 * math.Pow(10, 6)), "1.5 MB"},
		"format gigabytes over": {uint64(1.51 * math.Pow(10, 9)), "1.5 GB"},
		"format petabytes over": {uint64(1.51 * math.Pow(10, 12)), "1.5 TB"},
		"format terabytes over": {uint64(1.51 * math.Pow(10, 15)), "1.5 PB"},

		"format kilobytes exact": {uint64(1.3 * math.Pow(10, 3)), "1.3 KB"},
		"format megabytes exact": {uint64(1.3 * math.Pow(10, 6)), "1.3 MB"},
		"format gigabytes exact": {uint64(1.3 * math.Pow(10, 9)), "1.3 GB"},
		"format petabytes exact": {uint64(1.3 * math.Pow(10, 12)), "1.3 TB"},
		"format terabytes exact": {uint64(1.3 * math.Pow(10, 15)), "1.3 PB"},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestParseByteUnit(t *testing.T) {
	for desc, c := range map[string]*ParseByteUnitTestCase{
		"parse byte":     {"B", uint64(math.Pow(2, 0)), ""},
		"parse kibibyte": {"KIB", uint64(math.Pow(2, 10)), ""},
		"parse mebibyte": {"MIB", uint64(math.Pow(2, 20)), ""},
		"parse gibibyte": {"GIB", uint64(math.Pow(2, 30)), ""},
		"parse tebibyte": {"TIB", uint64(math.Pow(2, 40)), ""},
		"parse pebibyte": {"PIB", uint64(math.Pow(2, 50)), ""},

		"parse byte (lowercase)":     {"b", uint64(math.Pow(2, 0)), ""},
		"parse kibibyte (lowercase)": {"kib", uint64(math.Pow(2, 10)), ""},
		"parse mebibyte (lowercase)": {"mib", uint64(math.Pow(2, 20)), ""},
		"parse gibibyte (lowercase)": {"gib", uint64(math.Pow(2, 30)), ""},
		"parse tebibyte (lowercase)": {"tib", uint64(math.Pow(2, 40)), ""},
		"parse pebibyte (lowercase)": {"pib", uint64(math.Pow(2, 50)), ""},

		"parse byte (with space)":     {" B", uint64(math.Pow(2, 0)), ""},
		"parse kibibyte (with space)": {" KIB", uint64(math.Pow(2, 10)), ""},
		"parse mebibyte (with space)": {" MIB", uint64(math.Pow(2, 20)), ""},
		"parse gibibyte (with space)": {" GIB", uint64(math.Pow(2, 30)), ""},
		"parse tebibyte (with space)": {" TIB", uint64(math.Pow(2, 40)), ""},
		"parse pebibyte (with space)": {" PIB", uint64(math.Pow(2, 50)), ""},

		"parse byte (with space, lowercase)":     {" b", uint64(math.Pow(2, 0)), ""},
		"parse kibibyte (with space, lowercase)": {" kib", uint64(math.Pow(2, 10)), ""},
		"parse mebibyte (with space, lowercase)": {" mib", uint64(math.Pow(2, 20)), ""},
		"parse gibibyte (with space, lowercase)": {" gib", uint64(math.Pow(2, 30)), ""},
		"parse tebibyte (with space, lowercase)": {" tib", uint64(math.Pow(2, 40)), ""},
		"parse pebibyte (with space, lowercase)": {" pib", uint64(math.Pow(2, 50)), ""},

		"parse kilobyte": {"KB", uint64(math.Pow(10, 3)), ""},
		"parse megabyte": {"MB", uint64(math.Pow(10, 6)), ""},
		"parse gigabyte": {"GB", uint64(math.Pow(10, 9)), ""},
		"parse terabyte": {"TB", uint64(math.Pow(10, 12)), ""},
		"parse petabyte": {"PB", uint64(math.Pow(10, 15)), ""},

		"parse kilobyte (lowercase)": {"kb", uint64(math.Pow(10, 3)), ""},
		"parse megabyte (lowercase)": {"mb", uint64(math.Pow(10, 6)), ""},
		"parse gigabyte (lowercase)": {"gb", uint64(math.Pow(10, 9)), ""},
		"parse terabyte (lowercase)": {"tb", uint64(math.Pow(10, 12)), ""},
		"parse petabyte (lowercase)": {"pb", uint64(math.Pow(10, 15)), ""},

		"parse kilobyte (with space)": {" KB", uint64(math.Pow(10, 3)), ""},
		"parse megabyte (with space)": {" MB", uint64(math.Pow(10, 6)), ""},
		"parse gigabyte (with space)": {" GB", uint64(math.Pow(10, 9)), ""},
		"parse terabyte (with space)": {" TB", uint64(math.Pow(10, 12)), ""},
		"parse petabyte (with space)": {" PB", uint64(math.Pow(10, 15)), ""},

		"parse kilobyte (with space, lowercase)": {"kb", uint64(math.Pow(10, 3)), ""},
		"parse megabyte (with space, lowercase)": {"mb", uint64(math.Pow(10, 6)), ""},
		"parse gigabyte (with space, lowercase)": {"gb", uint64(math.Pow(10, 9)), ""},
		"parse terabyte (with space, lowercase)": {"tb", uint64(math.Pow(10, 12)), ""},
		"parse petabyte (with space, lowercase)": {"pb", uint64(math.Pow(10, 15)), ""},

		"parse unknown unit": {"imag", 0, "unknown unit: \"imag\""},
	} {
		t.Run(desc, c.Assert)
	}
}

func TestFormatBytesUnit(t *testing.T) {
	for desc, c := range map[string]*FormatBytesUnitTestCase{
		"format bytes":     {uint64(1 * math.Pow(10, 0)), humanize.Byte, "1"},
		"format kilobytes": {uint64(1 * math.Pow(10, 3)), humanize.Byte, "1000"},
		"format megabytes": {uint64(1 * math.Pow(10, 6)), humanize.Byte, "1000000"},
		"format gigabytes": {uint64(1 * math.Pow(10, 9)), humanize.Byte, "1000000000"},
		"format petabytes": {uint64(1 * math.Pow(10, 12)), humanize.Byte, "1000000000000"},
		"format terabytes": {uint64(1 * math.Pow(10, 15)), humanize.Byte, "1000000000000000"},

		"format kilobytes under": {uint64(1.49 * math.Pow(10, 3)), humanize.Byte, "1490"},
		"format megabytes under": {uint64(1.49 * math.Pow(10, 6)), humanize.Byte, "1490000"},
		"format gigabytes under": {uint64(1.49 * math.Pow(10, 9)), humanize.Byte, "1490000000"},
		"format petabytes under": {uint64(1.49 * math.Pow(10, 12)), humanize.Byte, "1490000000000"},
		"format terabytes under": {uint64(1.49 * math.Pow(10, 15)), humanize.Byte, "1490000000000000"},

		"format kilobytes over": {uint64(1.51 * math.Pow(10, 3)), humanize.Byte, "1510"},
		"format megabytes over": {uint64(1.51 * math.Pow(10, 6)), humanize.Byte, "1510000"},
		"format gigabytes over": {uint64(1.51 * math.Pow(10, 9)), humanize.Byte, "1510000000"},
		"format petabytes over": {uint64(1.51 * math.Pow(10, 12)), humanize.Byte, "1510000000000"},
		"format terabytes over": {uint64(1.51 * math.Pow(10, 15)), humanize.Byte, "1510000000000000"},
	} {
		t.Run(desc, c.Assert)
	}
}
