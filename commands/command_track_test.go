package commands

import "testing"

type checker struct {
	*testing.T
}

func TestAbsRelPath(t *testing.T) {
	c := &checker{t}

	ae := "/home/test"
	re := "test"
	a, r := absRelPath("test", "/home")
	c.Check(a, ae)
	c.Check(r, re)

	ae = "/home/test/some/directory/with/a/file.txt"
	re = "some/directory/with/a/file.txt"
	a, r = absRelPath("/home/test/some/directory/with/a/file.txt", "/home/test")
	c.Check(a, ae)
	c.Check(r, re)

	ae = "/home/test/some/directory/with/a/file.txt"
	re = "some/directory/with/a/file.txt"
	a, r = absRelPath("some/directory/with/a/file.txt", "/home/test")
	c.Check(a, ae)
	c.Check(r, re)
}

func (c *checker) Check(g, e string) {
	if g != e {
		c.Errorf("Expected %s got %s", e, g)
	}
}
