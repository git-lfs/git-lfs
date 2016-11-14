package lfs

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCatFileBatchValidInput(t *testing.T) {
	outc := make(chan *WrappedPointer)
	inc := make(chan string, 2)
	errc := make(chan error)
	shut1 := make(chan bool)
	shut2 := make(chan bool)

	go func(t *testing.T, shut chan bool, errc chan error) {
		for err := range errc {
			t.Errorf("err channel: %+v", err)
		}
		shut <- true
	}(t, shut1, errc)

	go func(t *testing.T, shut chan bool, outc chan *WrappedPointer) {
		expected := []*WrappedPointer{
			&WrappedPointer{
				Sha1: "60c8d8ab2adcf57a391163a7eeb0cdb8bf348e44",
				Size: 12345,
				Pointer: &Pointer{
					Oid:  "4d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393",
					Size: 12345,
				},
			},
			&WrappedPointer{
				Sha1: "e71d7db5669ed8dda17b4b1dceb30cb14745c591",
				Size: 12347,
				Pointer: &Pointer{
					Oid:  "7d7a214614ab2935c943f9e0ff69d22eadbb8f32b1258daaa5e2ca24d17e2393",
					Size: 12347,
				},
			},
		}

		i := 0
		t.Log("reading output in test")
		for actual := range outc {
			t.Logf("found actual: %+v", actual)
			if len(expected) <= i {
				t.Errorf("Cannot access index %d of output", i)
				break
			}

			assert.Equal(t, expected[i].Sha1, actual.Sha1)
			assert.Equal(t, expected[i].Oid, actual.Oid)
			assert.Equal(t, expected[i].Size, actual.Size)
			i++
		}

		if len(expected) > i {
			t.Errorf("got to index %d of %d", i, len(expected))
		}

		shut <- true
	}(t, shut2, outc)

	if err := runCatFileBatch(outc, inc, errc); err != nil {
		t.Fatal(err)
	}

	t.Logf("sending input")
	inc <- "126fd41019b623ce1621a638d2535b26e0edb4df"
	inc <- "60c8d8ab2adcf57a391163a7eeb0cdb8bf348e44"
	time.Sleep(1 * time.Second)
	inc <- "e71d7db5669ed8dda17b4b1dceb30cb14745c591"
	close(inc)

	<-shut1
	<-shut2
}
