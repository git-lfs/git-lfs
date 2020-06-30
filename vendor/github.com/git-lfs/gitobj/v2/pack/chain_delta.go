package pack

import "fmt"

// ChainDelta represents a "delta" component of a delta-base chain.
type ChainDelta struct {
	// Base is the base delta-base chain that this delta should be applied
	// to. It can be a ChainBase in the simple case, or it can itself be a
	// ChainDelta, which resolves against another ChainBase, when the
	// delta-base chain is of length greater than 2.
	base Chain
	// delta is the set of copy/add instructions to apply on top of the
	// base.
	delta []byte
}

// Unpack applies the delta operation to the previous delta-base chain, "base".
//
// If any of the delta-base instructions were invalid, an error will be
// returned.
func (d *ChainDelta) Unpack() ([]byte, error) {
	base, err := d.base.Unpack()
	if err != nil {
		return nil, err
	}

	return patch(base, d.delta)
}

// Type returns the type of the base of the delta-base chain.
func (d *ChainDelta) Type() PackedObjectType {
	return d.base.Type()
}

// patch applies the delta instructions in "delta" to the base given as "base".
// It returns the result of applying those patch instructions to base, but does
// not modify base itself.
//
// If any of the delta instructions were malformed, or otherwise could not be
// applied to the given base, an error will returned, along with an empty set of
// data.
func patch(base, delta []byte) ([]byte, error) {
	srcSize, pos := patchDeltaHeader(delta, 0)
	if srcSize != int64(len(base)) {
		// The header of the delta gives the size of the source contents
		// that it is a patch over.
		//
		// If this does not match with the srcSize, return an error
		// early so as to avoid a possible bounds error below.
		return nil, fmt.Errorf("gitobj/pack: invalid delta data")
	}

	// The remainder of the delta header contains the destination size, and
	// moves the "pos" offset to the correct position to begin the set of
	// delta instructions.
	destSize, pos := patchDeltaHeader(delta, pos)

	dest := make([]byte, 0, destSize)

	for pos < len(delta) {
		c := int(delta[pos])
		pos += 1

		if c&0x80 != 0 {
			// If the most significant bit (MSB, at position 0x80)
			// is set, this is a copy instruction. Advance the
			// position one byte backwards, and initialize variables
			// for the copy offset and size instructions.
			pos -= 1

			var co, cs int

			// The lower-half of "c" (0000 1111) defines a "bitmask"
			// for the copy offset.
			if c&0x1 != 0 {
				pos += 1
				co = int(delta[pos])
			}
			if c&0x2 != 0 {
				pos += 1
				co |= (int(delta[pos]) << 8)
			}
			if c&0x4 != 0 {
				pos += 1
				co |= (int(delta[pos]) << 16)
			}
			if c&0x8 != 0 {
				pos += 1
				co |= (int(delta[pos]) << 24)
			}

			// The upper-half of "c" (1111 0000) defines a "bitmask"
			// for the size of the copy instruction.
			if c&0x10 != 0 {
				pos += 1
				cs = int(delta[pos])
			}
			if c&0x20 != 0 {
				pos += 1
				cs |= (int(delta[pos]) << 8)
			}
			if c&0x40 != 0 {
				pos += 1
				cs |= (int(delta[pos]) << 16)
			}

			if cs == 0 {
				// If the copy size is zero, we assume that it
				// is the next whole number after the max uint32
				// value.
				cs = 0x10000
			}
			pos += 1

			// Once we have the copy offset and length defined, copy
			// that number of bytes from the base into the
			// destination. Since we are copying from the base and
			// not the delta, the position into the delta ("pos")
			// need not be updated.
			dest = append(dest, base[co:co+cs]...)
		} else if c != 0 {
			// If the most significant bit (MSB) is _not_ set, we
			// instead process a copy instruction, where "c" is the
			// number of successive bytes in the delta patch to add
			// to the output.
			//
			// Copy the bytes and increment the read pointer
			// forward.
			dest = append(dest, delta[pos:int(pos)+c]...)

			pos += int(c)
		} else {
			// Otherwise, "c" is 0, and is an invalid delta
			// instruction.
			//
			// Return immediately.
			return nil, fmt.Errorf(
				"gitobj/pack: invalid delta data")
		}
	}

	if destSize != int64(len(dest)) {
		// If after patching the delta against the base, the destination
		// size is different than the expected destination size, we have
		// an invalid set of patch instructions.
		//
		// Return immediately.
		return nil, fmt.Errorf("gitobj/pack: invalid delta data")
	}
	return dest, nil
}

// patchDeltaHeader examines the header within delta at the given offset, and
// returns the size encoded within it, as well as the ending offset where begins
// the next header, or the patch instructions.
func patchDeltaHeader(delta []byte, pos int) (size int64, end int) {
	var shift uint
	var c int64

	for shift == 0 || c&0x80 != 0 {
		if len(delta) <= pos {
			panic("gitobj/pack: invalid delta header")
		}

		c = int64(delta[pos])

		pos++
		size |= (c & 0x7f) << shift
		shift += 7
	}

	return size, pos
}
