package tools

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderedSetAddAddsElements(t *testing.T) {
	s := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.False(t, s.Contains("d"),
		"tools: did not expected s to contain \"d\"")

	assert.True(t, s.Add("d"))

	assert.True(t, s.Contains("d"),
		"tools: expected s to contain \"d\"")
}

func TestOrderedSetContainsReturnsTrueForItemsItContains(t *testing.T) {
	s := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.True(t, s.Contains("b"),
		"tools: expected s to contain element \"b\"")
}

func TestOrderedSetContainsReturnsFalseForItemsItDoesNotContains(t *testing.T) {
	s := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.False(t, s.Contains("d"),
		"tools: did not expect s to contain element \"d\"")
}

func TestOrderedSetContainsAllReturnsTrueWhenAllElementsAreContained(t *testing.T) {
	s := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.True(t, s.ContainsAll("b", "c"),
		"tools: expected s to contain element \"b\" and \"c\"")
}

func TestOrderedSetContainsAllReturnsFalseWhenAllElementsAreNotContained(t *testing.T) {
	s := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.False(t, s.ContainsAll("b", "c", "d"),
		"tools: did not expect s to contain element \"b\", \"c\" and \"d\"")
}

func TestOrderedSetIsSubsetReturnsTrueWhenOtherContainsAllElements(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b", "c"})
	s2 := NewOrderedSetFromSlice([]string{"a", "b"})

	assert.True(t, s1.IsSubset(s2),
		"tools: expected [a, b] to be a subset of [a, b, c]")
}

func TestOrderedSetIsSubsetReturnsFalseWhenOtherDoesNotContainAllElements(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})
	s2 := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.False(t, s1.IsSubset(s2),
		"tools: did not expect [a, b, c] to be a subset of [a, b]")
}

func TestOrderedSetIsSupersetReturnsTrueWhenContainsAllElementsOfOther(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})
	s2 := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.True(t, s1.IsSuperset(s2),
		"tools: expected [a, b, c] to be a superset of [a, b]")
}

func TestOrderedSetIsSupersetReturnsFalseWhenDoesNotContainAllElementsOfOther(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b", "c"})
	s2 := NewOrderedSetFromSlice([]string{"a", "b"})

	assert.False(t, s1.IsSuperset(s2),
		"tools: did not expect [a, b] to be a superset of [a, b, c]")
}

func TestOrderedSetUnion(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a"})
	s2 := NewOrderedSetFromSlice([]string{"b", "a"})

	elems := make([]string, 0)
	for e := range s1.Union(s2).Iter() {
		elems = append(elems, e)
	}

	require.Len(t, elems, 2)
	assert.Equal(t, "a", elems[0])
	assert.Equal(t, "b", elems[1])
}

func TestOrderedSetIntersect(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a"})
	s2 := NewOrderedSetFromSlice([]string{"b", "a"})

	elems := make([]string, 0)
	for e := range s1.Intersect(s2).Iter() {
		elems = append(elems, e)
	}

	require.Len(t, elems, 1)
	assert.Equal(t, "a", elems[0])
}

func TestOrderedSetDifference(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})
	s2 := NewOrderedSetFromSlice([]string{"a"})

	elems := make([]string, 0)
	for e := range s1.Difference(s2).Iter() {
		elems = append(elems, e)
	}

	require.Len(t, elems, 1)
	assert.Equal(t, "b", elems[0])
}

func TestOrderedSetSymmetricDifference(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})
	s2 := NewOrderedSetFromSlice([]string{"b", "c"})

	elems := make([]string, 0)
	for e := range s1.SymmetricDifference(s2).Iter() {
		elems = append(elems, e)
	}

	require.Len(t, elems, 2)
	assert.Equal(t, "a", elems[0])
	assert.Equal(t, "c", elems[1])
}

func TestOrderedSetClear(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})

	assert.Equal(t, 2, s1.Cardinality())

	s1.Clear()

	assert.Equal(t, 0, s1.Cardinality())
}

func TestOrderedSetRemove(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})

	assert.True(t, s1.Contains("a"), "tools: expected [a, b] to contain 'a'")
	assert.True(t, s1.Contains("b"), "tools: expected [a, b] to contain 'b'")

	s1.Remove("a")

	assert.False(t, s1.Contains("a"), "tools: did not expect to find 'a' in [b]")
	assert.True(t, s1.Contains("b"), "tools: expected [b] to contain 'b'")
}

func TestOrderedSetCardinality(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b"})

	assert.Equal(t, 2, s1.Cardinality(),
		"tools: expected cardinality of [a, b] to equal 2")
}

func TestOrderedSetIter(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	elems := make([]string, 0)
	for e := range s1.Iter() {
		elems = append(elems, e)
	}

	require.Len(t, elems, 3)
	assert.Equal(t, "a", elems[0])
	assert.Equal(t, "b", elems[1])
	assert.Equal(t, "c", elems[2])
}

func TestOrderedSetEqualReturnsTrueWhenSameElementsInSameOrder(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b", "c"})
	s2 := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	assert.True(t, s1.Equal(s2),
		"tools: expected [a, b, c] to equal [a, b, c]")
}

func TestOrderedSetEqualReturnsFalseWhenSameElementsInDifferentOrder(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b", "c"})
	s2 := NewOrderedSetFromSlice([]string{"a", "c", "b"})

	assert.False(t, s1.Equal(s2),
		"tools: did not expect [a, b, c] to equal [a, c, b]")
}

func TestOrderedSetEqualReturnsFalseWithDifferentCardinalities(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a"})
	s2 := NewOrderedSetFromSlice([]string{"a", "b"})

	assert.False(t, s1.Equal(s2),
		"tools: did not expect [a] to equal [a, b]")
}

func TestOrderedSetClone(t *testing.T) {
	s1 := NewOrderedSetFromSlice([]string{"a", "b", "c"})

	s2 := s1.Clone()

	elems := make([]string, 0)
	for e := range s2.Iter() {
		elems = append(elems, e)
	}

	require.Len(t, elems, 3)
	assert.Equal(t, "a", elems[0])
	assert.Equal(t, "b", elems[1])
	assert.Equal(t, "c", elems[2])
}
