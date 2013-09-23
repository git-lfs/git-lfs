package gitmedia

import "fmt"

func init() {
	registerCommand("version", NewCommand(func(c *Command) {
		comics := c.FlagSet.Bool("comics", false, "easter egg")
		c.parse()

		if *comics {
			fmt.Println("Nothing may see Gah Lak Tus and survive.")
		} else {
			fmt.Printf("git-media version %s\n", Version)
		}
	}))
}
