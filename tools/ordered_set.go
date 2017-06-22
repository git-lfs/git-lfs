package tools

// OrderedSet is a unique set of strings that maintains insertion order.
type OrderedSet struct {
	// s is the set of strings that we're keeping track of.
	s []string
	// m is a mapping of string value "s" into the index "i" that that
	// string is present in in the given "s".
	m map[string]int
}

// NewOrderedSet creates an ordered set with no values.
func NewOrderedSet() *OrderedSet {
	return NewOrderedSetWithCapacity(0)
}

// NewOrderedSetWithCapacity creates a new ordered set with no values. The
// returned ordered set can be appended to "capacity" number of times before it
// grows internally.
func NewOrderedSetWithCapacity(capacity int) *OrderedSet {
	return &OrderedSet{
		s: make([]string, 0, capacity),
		m: make(map[string]int, capacity),
	}
}

// NewOrderedSetFromSlice returns a new ordered set with the elements given in
// the slice "s".
func NewOrderedSetFromSlice(s []string) *OrderedSet {
	set := NewOrderedSetWithCapacity(len(s))
	for _, e := range s {
		set.Add(e)
	}

	return set
}

// Add adds the given element "i" to the ordered set, unless the element is
// already present. It returns whether or not the element was added.
func (s *OrderedSet) Add(i string) bool {
	if _, ok := s.m[i]; ok {
		return false
	}

	s.s = append(s.s, i)
	s.m[i] = len(s.s) - 1

	return true
}

// Contains returns whether or not the given "i" is contained in this ordered
// set. It is a constant-time operation.
func (s *OrderedSet) Contains(i string) bool {
	if _, ok := s.m[i]; ok {
		return true
	}
	return false
}

// ContainsAll returns whether or not all of the given items in "i" are present
// in the ordered set.
func (s *OrderedSet) ContainsAll(i ...string) bool {
	for _, item := range i {
		if !s.Contains(item) {
			return false
		}
	}
	return true
}

// IsSubset returns whether other is a subset of this ordered set. In other
// words, it returns whether or not all of the elements in "other" are also
// present in this set.
func (s *OrderedSet) IsSubset(other *OrderedSet) bool {
	for _, i := range other.s {
		if !s.Contains(i) {
			return false
		}
	}
	return true
}

// IsSuperset returns whether or not this set is a superset of "other". In other
// words, it returns whether or not all of the elements in this set are also in
// the set "other".
func (s *OrderedSet) IsSuperset(other *OrderedSet) bool {
	return other.IsSubset(s)
}

// Union returns a union of this set with the given set "other". It returns the
// items that are in either set while maintaining uniqueness constraints. It
// preserves ordered within each set, and orders the elements in this set before
// the elements in "other".
//
// It is an O(n+m) operation.
func (s *OrderedSet) Union(other *OrderedSet) *OrderedSet {
	union := NewOrderedSetWithCapacity(other.Cardinality() + s.Cardinality())

	for _, e := range s.s {
		union.Add(e)
	}
	for _, e := range other.s {
		union.Add(e)
	}

	return union
}

// Intersect returns the elements that are in both this set and then given
// "ordered" set. It is an O(min(n, m)) (in other words, O(n)) operation.
func (s *OrderedSet) Intersect(other *OrderedSet) *OrderedSet {
	intersection := NewOrderedSetWithCapacity(MinInt(
		s.Cardinality(), other.Cardinality()))

	if s.Cardinality() < other.Cardinality() {
		for _, elem := range s.s {
			if other.Contains(elem) {
				intersection.Add(elem)
			}
		}
	} else {
		for _, elem := range other.s {
			if s.Contains(elem) {
				intersection.Add(elem)
			}
		}
	}

	return intersection
}

// Difference returns the elements that are in this set, but not included in
// other.
func (s *OrderedSet) Difference(other *OrderedSet) *OrderedSet {
	diff := NewOrderedSetWithCapacity(s.Cardinality())
	for _, e := range s.s {
		if !other.Contains(e) {
			diff.Add(e)
		}
	}

	return diff
}

// SymmetricDifference returns the elements that are not present in both sets.
func (s *OrderedSet) SymmetricDifference(other *OrderedSet) *OrderedSet {
	left := s.Difference(other)
	right := other.Difference(s)

	return left.Union(right)
}

// Clear removes all elements from this set.
func (s *OrderedSet) Clear() {
	s.s = make([]string, 0)
	s.m = make(map[string]int, 0)
}

// Remove removes the given element "i" from this set.
func (s *OrderedSet) Remove(i string) {
	idx, ok := s.m[i]
	if !ok {
		return
	}

	rest := MinInt(idx+1, len(s.s)-1)

	s.s = append(s.s[:idx], s.s[rest:]...)
	for _, e := range s.s[rest:] {
		s.m[e] = s.m[e] - 1
	}
	delete(s.m, i)
}

// Cardinality returns the cardinality of this set.
func (s *OrderedSet) Cardinality() int {
	return len(s.s)
}

// Iter returns a channel which yields the elements in this set in insertion
// order.
func (s *OrderedSet) Iter() <-chan string {
	c := make(chan string)
	go func() {
		for _, i := range s.s {
			c <- i
		}
		close(c)
	}()

	return c
}

// Equal returns whether this element has the same number, identity and ordering
// elements as given in "other".
func (s *OrderedSet) Equal(other *OrderedSet) bool {
	if s.Cardinality() != other.Cardinality() {
		return false
	}

	for e, i := range s.m {
		if ci, ok := other.m[e]; !ok || ci != i {
			return false
		}
	}

	return true
}

// Clone returns a deep copy of this set.
func (s *OrderedSet) Clone() *OrderedSet {
	clone := NewOrderedSetWithCapacity(s.Cardinality())
	for _, i := range s.s {
		clone.Add(i)
	}
	return clone
}
