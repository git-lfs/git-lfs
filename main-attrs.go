package main

import (
	"fmt"

	"github.com/git-lfs/git-lfs/git/gitattributes"
)

const file = `
a[[:space:]]*.dat filter=lfs diff=lfs merge=lfs -text
b[[:space:]]*.dat filter=lfs diff=lfs merge=lfs -text
c[[:space:]]*.dat filter=lfs diff=lfs merge=lfs -text
d[[:space:]]*.dat filter=lfs diff=lfs merge=lfs -text
`

func main() {
	t := &gitattributes.Tree{Repo: "/Users/ttaylorr/Desktop/example"}
	attrs, err := t.Attributes("a/1.dat")

	fmt.Println(attrs)
	for k, v := range attrs {
		fmt.Printf("%s=%s\n", k, v)
	}

	if err != nil {
		panic(err)
	}

	// entries, err := gitattributes.ParseEntries("/", strings.NewReader(file))

	// fmt.Printf("entries=%+v, err=%+v\n", entries, err)
	// for _, entry := range entries.Entries {
	// 	fmt.Printf("\tentry=%+v\n", entry)
	// }
}
