package main

import (
	"fmt"
	"go/build"
	"os"
	"strings"
)

var packages map[string]string = make(map[string]string)

func generate_target(srcdir string, pkgdir string, prefix string, ctx build.Context) string {
	pkg, _ := ctx.ImportDir(srcdir+pkgdir, 0)
	name := pkg.Name
	var deps []string
	for _, imp := range pkg.Imports {
		if strings.HasPrefix(imp, prefix) {
			imp = strings.TrimPrefix(imp, prefix)
			if packages[imp] == "" {
				packages[imp] = generate_target(srcdir, imp, prefix, ctx)
			}
			deps = append(deps, "$(LIBS_"+packages[imp]+")")
		}
	}
	if pkgdir != "" {
		fmt.Printf("SRCDIR_%s := $(SRCDIR)%s/\n", name, pkgdir)
	} else {
		fmt.Printf("SRCDIR_%s := $(SRCDIR)\n", name)
	}
	fmt.Printf("SRC_%s := $(addprefix $(SRCDIR_%s), %s)\n", name, name, strings.Join(pkg.GoFiles, " "))
	fmt.Printf("DEPS_%s := %s\n", name, strings.Join(deps, " "))
	if pkgdir != "" {
		fmt.Printf("OBJ_%s := $(LIBDIR)/%s.o\n", name, pkgdir)
		fmt.Printf("LIB_%s := $(LIBDIR)/%s.a\n", name, pkgdir)
		fmt.Printf("LIBS_%s := $(LIB_%s) $(DEPS_%s)\n", name, name, name)
		fmt.Printf("$(OBJ_%s) : $(SRC_%s) $(DEPS_%s)\n", name, name, name)
		fmt.Printf("\t@mkdir -p $(dir $@)\n")
		fmt.Printf("\t$(GOC) $(GOFLAGS) -c -o $@ $(SRC_%s)\n", name)
	}
	return name
}

func main() {
	srcdir := os.Args[1]
	prefix := os.Args[2]
	ctx := build.Default
	ctx.CgoEnabled = false
	generate_target(srcdir, "", prefix, ctx)
}
