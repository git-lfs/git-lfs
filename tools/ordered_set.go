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

// Clone returns a deep copy of this set.
func (s *OrderedSet) Clone() *OrderedSet {
	clone := NewOrderedSetWithCapacity(s.Cardinality())
	for _, i := range s.s {
		clone.Add(i)
	}
	return clone
}
