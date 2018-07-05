package pack

import "fmt"

// bounds encapsulates the window of search for a single iteration of binary
// search.
//
// Callers may choose to treat the return values from Left() and Right() as
// inclusive or exclusive. *bounds makes no assumptions on the inclusivity of
// those values.
//
// See: *gitobj/pack.Index for more.
type bounds struct {
	// left is the left or lower bound of the bounds.
	left int64
	// right is the rightmost or upper bound of the bounds.
	right int64
}

// newBounds returns a new *bounds instance with the given left and right
// values.
func newBounds(left, right int64) *bounds {
	return &bounds{
		left:  left,
		right: right,
	}
}

// Left returns the leftmost value or lower bound of this *bounds instance.
func (b *bounds) Left() int64 {
	return b.left
}

// right returns the rightmost value or upper bound of this *bounds instance.
func (b *bounds) Right() int64 {
	return b.right
}

// WithLeft returns a new copy of this *bounds instance, replacing the left
// value with the given argument.
func (b *bounds) WithLeft(new int64) *bounds {
	return &bounds{
		left:  new,
		right: b.right,
	}
}

// WithRight returns a new copy of this *bounds instance, replacing the right
// value with the given argument.
func (b *bounds) WithRight(new int64) *bounds {
	return &bounds{
		left:  b.left,
		right: new,
	}
}

// Equal returns whether or not the receiving *bounds instance is equal to the
// given one:
//
//   - If both the argument and receiver are nil, they are given to be equal.
//   - If both the argument and receiver are not nil, and they share the same
//     Left() and Right() values, they are equal.
//   - If both the argument and receiver are not nil, but they do not share the
//     same Left() and Right() values, they are not equal.
//   - If either the argument or receiver is nil, but the other is not, they are
//     not equal.
func (b *bounds) Equal(other *bounds) bool {
	if b == nil {
		if other == nil {
			return true
		}
		return false
	}

	if other == nil {
		return false
	}

	return b.left == other.left &&
		b.right == other.right
}

// String returns a string representation of this bounds instance, given as:
//
//   [<left>,<right>]
func (b *bounds) String() string {
	return fmt.Sprintf("[%d,%d]", b.Left(), b.Right())
}
