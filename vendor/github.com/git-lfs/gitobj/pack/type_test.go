package pack

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type PackedObjectStringTestCase struct {
	T PackedObjectType

	Expected string
	Panic    bool
}

func (c *PackedObjectStringTestCase) Assert(t *testing.T) {
	if c.Panic {
		defer func() {
			err := recover()

			if err == nil {
				t.Fatalf("gitobj/pack: expected panic()")
			}

			assert.Equal(t, c.Expected, fmt.Sprintf("%s", err))
		}()
	}

	assert.Equal(t, c.Expected, c.T.String())
}

func TestPackedObjectTypeString(t *testing.T) {
	for desc, c := range map[string]*PackedObjectStringTestCase{
		"TypeNone": {T: TypeNone, Expected: "<none>"},

		"TypeCommit": {T: TypeCommit, Expected: "commit"},
		"TypeTree":   {T: TypeTree, Expected: "tree"},
		"TypeBlob":   {T: TypeBlob, Expected: "blob"},
		"TypeTag":    {T: TypeTag, Expected: "tag"},

		"TypeObjectOffsetDelta": {T: TypeObjectOffsetDelta,
			Expected: "obj_ofs_delta"},
		"TypeObjectReferenceDelta": {T: TypeObjectReferenceDelta,
			Expected: "obj_ref_delta"},

		"unknown type": {T: PackedObjectType(5), Panic: true,
			Expected: "gitobj/pack: unknown object type: 5"},
	} {
		t.Run(desc, c.Assert)
	}
}
