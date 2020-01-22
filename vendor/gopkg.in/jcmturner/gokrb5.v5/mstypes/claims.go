package mstypes

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"gopkg.in/jcmturner/rpc.v0/ndr"
)

// Compression format assigned numbers.
const (
	CompressionFormatNone       uint16 = 0
	CompressionFormatLZNT1      uint16 = 2
	CompressionFormatXPress     uint16 = 3
	CompressionFormatXPressHuff uint16 = 4
)

// ClaimsSourceType
const ClaimsSourceTypeAD uint16 = 1

// Claim Type assigned numbers
const (
	ClaimTypeIDInt64    uint16 = 1
	ClaimTypeIDUInt64   uint16 = 2
	ClaimTypeIDString   uint16 = 3
	ClaimsTypeIDBoolean uint16 = 6
)

// ClaimsBlob implements https://msdn.microsoft.com/en-us/library/hh554119.aspx
type ClaimsBlob struct {
	Size        uint32
	EncodedBlob []byte
}

// ReadClaimsBlob reads a ClaimsBlob from the byte slice.
func ReadClaimsBlob(b *[]byte, p *int, e *binary.ByteOrder) (c ClaimsBlob) {
	c.Size = ndr.ReadUint32(b, p, e)
	c.EncodedBlob = ndr.ReadBytes(b, p, int(c.Size), e)
	return
}

// ClaimsSetMetadata implements https://msdn.microsoft.com/en-us/library/hh554073.aspx
type ClaimsSetMetadata struct {
	claimsSetSize             uint32
	ClaimsSet                 ClaimsSet
	CompressionFormat         uint16 // Enum see constants for options
	uncompressedClaimsSetSize uint32
	ReservedType              uint16
	reservedFieldSize         uint32
	ReservedField             []byte
}

// ClaimSet implements https://msdn.microsoft.com/en-us/library/hh554122.aspx
type ClaimsSet struct {
	ClaimsArrayCount  uint32
	ClaimsArrays      []ClaimsArray
	ReservedType      uint16
	reservedFieldSize uint32
	ReservedField     []byte
}

// ClaimsArray implements https://msdn.microsoft.com/en-us/library/hh536458.aspx
type ClaimsArray struct {
	ClaimsSourceType uint16
	ClaimsCount      uint32
	ClaimsEntries    []ClaimEntry
}

// ClaimEntry implements https://msdn.microsoft.com/en-us/library/hh536374.aspx
type ClaimEntry struct {
	ID         string //utf16string
	Type       uint16 // enums are 16 bit https://msdn.microsoft.com/en-us/library/windows/desktop/aa366818(v=vs.85).aspx
	TypeInt64  ClaimTypeInt64
	TypeUInt64 ClaimTypeUInt64
	TypeString ClaimTypeString
	TypeBool   ClaimTypeBoolean
}

// ClaimTypeInt64 is a claim of type int64
type ClaimTypeInt64 struct {
	ValueCount uint32
	Value      []int64
}

// ClaimTypeUInt64 is a claim of type uint64
type ClaimTypeUInt64 struct {
	ValueCount uint32
	Value      []uint64
}

// ClaimTypeString is a claim of type string
type ClaimTypeString struct {
	ValueCount uint32
	Value      []string
}

// ClaimTypeBoolean is a claim of type bool
type ClaimTypeBoolean struct {
	ValueCount uint32
	Value      []bool
}

// ReadClaimsSetMetadata reads a ClaimsSetMetadata from the bytes slice.
func ReadClaimsSetMetadata(b *[]byte, p *int, e *binary.ByteOrder) (c ClaimsSetMetadata, err error) {
	c.claimsSetSize = ndr.ReadUint32(b, p, e)
	*p += 4 //Move over pointer to ClaimSet array
	c.CompressionFormat = ndr.ReadUint16(b, p, e)
	// TODO Currently compression is not supported so if it is compressed we just have to return.
	if c.CompressionFormat != CompressionFormatNone {
		*p = len(*b)
		return
	}
	c.uncompressedClaimsSetSize = ndr.ReadUint32(b, p, e)
	c.ReservedType = ndr.ReadUint16(b, p, e)
	c.reservedFieldSize = ndr.ReadUint32(b, p, e)
	*p += 4 //Move over pointer to ReservedField array
	var ah ndr.ConformantArrayHeader
	if c.claimsSetSize > 0 {
		// ClaimsSet is a conformant array https://msdn.microsoft.com/en-us/library/windows/desktop/aa373603(v=vs.85).aspx
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(b, p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(c.claimsSetSize) {
			err = errors.New("error with size of CLAIMS_SET array")
			return
		}
		csb := ndr.ReadBytes(b, p, int(c.claimsSetSize), e)
		//TODO put decompression here
		c.ClaimsSet, err = ReadClaimsSet(csb)
		if err != nil {
			return
		}
	}
	if c.reservedFieldSize > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(b, p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(c.reservedFieldSize) {
			err = errors.New("error with size of CLAIMS_SET_METADATA's reserved field array")
			return
		}
		c.ReservedField = ndr.ReadBytes(b, p, int(c.reservedFieldSize), e)
	}
	return
}

// ReadClaimsSet reads a ClaimsSet from the bytes slice.
func ReadClaimsSet(b []byte) (c ClaimsSet, err error) {
	ch, _, p, err := ndr.ReadHeaders(&b)
	if err != nil {
		err = fmt.Errorf("error parsing NDR byte stream headers of CLAIMS_SET: %v", err)
		return
	}
	e := &ch.Endianness
	//The next 4 bytes are an RPC unique pointer referent. We just skip these
	p += 4

	c.ClaimsArrayCount = ndr.ReadUint32(&b, &p, e)
	p += 4 //Move over pointer to claims array
	c.ReservedType = ndr.ReadUint16(&b, &p, e)
	c.reservedFieldSize = ndr.ReadUint32(&b, &p, e)
	p += 4 //Move over pointer to ReservedField array

	var ah ndr.ConformantArrayHeader
	if c.ClaimsArrayCount > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(&b, &p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(c.ClaimsArrayCount) {
			err = errors.New("error with size of CLAIMS_SET's claims array")
			return
		}
		c.ClaimsArrays = make([]ClaimsArray, c.ClaimsArrayCount, c.ClaimsArrayCount)
		for i := range c.ClaimsArrays {
			c.ClaimsArrays[i], err = ReadClaimsArray(&b, &p, e)
			if err != nil {
				return
			}
		}
	}
	if c.reservedFieldSize > 0 {
		ah, err = ndr.ReadUniDimensionalConformantArrayHeader(&b, &p, e)
		if err != nil {
			return
		}
		if ah.MaxCount != int(c.reservedFieldSize) {
			err = errors.New("error with size of CLAIMS_SET's reserved field array")
			return
		}
		c.ReservedField = ndr.ReadBytes(&b, &p, int(c.reservedFieldSize), e)
	}
	return c, nil
}

// ReadClaimsArray reads a ClaimsArray from the bytes slice.
func ReadClaimsArray(b *[]byte, p *int, e *binary.ByteOrder) (c ClaimsArray, err error) {
	c.ClaimsSourceType = ndr.ReadUint16(b, p, e)
	c.ClaimsCount = ndr.ReadUint32(b, p, e)
	*p += 4 //Move over pointer to claims array
	ah, err := ndr.ReadUniDimensionalConformantArrayHeader(b, p, e)
	if err != nil {
		return
	}
	if ah.MaxCount != int(c.ClaimsCount) {
		err = errors.New("error with size of CLAIMS_ARRAY's claims entries")
		return
	}
	c.ClaimsEntries = make([]ClaimEntry, c.ClaimsCount, c.ClaimsCount)
	for i := range c.ClaimsEntries {
		var vc uint32
		c.ClaimsEntries[i].Type, vc, err = ReadClaimEntriesUnionHeaders(b, p, e)
		if err != nil {
			return
		}
		switch c.ClaimsEntries[i].Type {
		case ClaimTypeIDInt64:
			c.ClaimsEntries[i].TypeInt64.ValueCount = vc
		case ClaimTypeIDUInt64:
			c.ClaimsEntries[i].TypeUInt64.ValueCount = vc
		case ClaimTypeIDString:
			c.ClaimsEntries[i].TypeString.ValueCount = vc
		case ClaimsTypeIDBoolean:
			c.ClaimsEntries[i].TypeBool.ValueCount = vc
		}
	}
	for i := range c.ClaimsEntries {
		err = FillClaimEntry(b, p, e, &c.ClaimsEntries[i])
		if err != nil {
			return
		}
	}
	return
}

func ReadClaimEntriesUnionHeaders(b *[]byte, p *int, e *binary.ByteOrder) (uint16, uint32, error) {
	*p += 4
	// This is an NDR union: http://pubs.opengroup.org/onlinepubs/9629399/chap14.htm#tagfcjh_39
	// The discriminant [tag] is marshalled into the transmitted data stream twice:
	// once as the field or parameter in the procedure argument list and
	// once as the first part of the union representation [value]
	t1 := ndr.ReadUint16(b, p, e)
	t2 := ndr.ReadUint16(b, p, e)
	if t1 != t2 {
		return 0, 0, ndr.Malformed{EText: "malformed NDR encoding of CLAIM_ENTRY union"}
	}
	vc := ndr.ReadUint32(b, p, e)
	*p += 4 //Move over pointer to array of values
	return t1, vc, nil
}

// FillClaimEntry reads a ClaimEntry from the bytes slice.
func FillClaimEntry(b *[]byte, p *int, e *binary.ByteOrder, c *ClaimEntry) (err error) {
	c.ID, err = ndr.ReadConformantVaryingString(b, p, e)
	if err != nil {
		return
	}
	ah, err := ndr.ReadUniDimensionalConformantArrayHeader(b, p, e)
	if err != nil {
		return
	}
	switch c.Type {
	case ClaimTypeIDInt64:
		if ah.MaxCount != int(c.TypeInt64.ValueCount) {
			return errors.New("error with size of CLAIM_ENTRY's value")
		}
		c.TypeInt64.Value = make([]int64, c.TypeInt64.ValueCount, c.TypeInt64.ValueCount)
		for i := range c.TypeInt64.Value {
			buf := bytes.NewReader((*b)[*p : *p+8])
			err = binary.Read(buf, *e, &c.TypeInt64.Value[i])
			if err != nil {
				return
			}
			*p += 8 // progress position for a uint64
		}
	case ClaimTypeIDUInt64:
		if ah.MaxCount != int(c.TypeUInt64.ValueCount) {
			return errors.New("error with size of CLAIM_ENTRY's value")
		}
		c.TypeUInt64.Value = make([]uint64, c.TypeUInt64.ValueCount, c.TypeUInt64.ValueCount)
		for i := range c.TypeUInt64.Value {
			c.TypeUInt64.Value[i] = ndr.ReadUint64(b, p, e)
		}
	case ClaimTypeIDString:
		if ah.MaxCount != int(c.TypeString.ValueCount) {
			return errors.New("error with size of CLAIM_ENTRY's value")
		}
		c.TypeString.Value = make([]string, c.TypeString.ValueCount, c.TypeString.ValueCount)
		*p += 4 * (int(c.TypeString.ValueCount)) // Move over pointers
		for i := range c.TypeString.Value {
			c.TypeString.Value[i], err = ndr.ReadConformantVaryingString(b, p, e)
			if err != nil {
				return
			}
		}
	case ClaimsTypeIDBoolean:
		if ah.MaxCount != int(c.TypeBool.ValueCount) {
			return errors.New("error with size of CLAIM_ENTRY's value")
		}
		c.TypeBool.Value = make([]bool, c.TypeBool.ValueCount, c.TypeBool.ValueCount)
		for i := range c.TypeBool.Value {
			if ndr.ReadUint64(b, p, e) != 0 {
				c.TypeBool.Value[i] = true
			}
		}
	}
	return
}
