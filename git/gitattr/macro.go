package gitattr

import (
	"fmt"
)

type MacroProcessor struct {
	macros map[string][]*Attr
}

// NewMacroProcessor returns a new MacroProcessor object for parsing macros.
func NewMacroProcessor() *MacroProcessor {
	macros := make(map[string][]*Attr)

	// This is built into Git.
	macros["binary"] = []*Attr{
		&Attr{K: "diff", V: "false"},
		&Attr{K: "merge", V: "false"},
		&Attr{K: "text", V: "false"},
	}

	return &MacroProcessor{
		macros: macros,
	}
}

// ProcessLines reads the specified lines, returning a new set of lines which
// all have a valid pattern.  If readMacros is true, it additionally loads any
// macro lines as it reads them.
func (mp *MacroProcessor) ProcessLines(lines []Line, readMacros bool) []PatternLine {
	fmt.Println("I0")
	result := make([]PatternLine, 0, len(lines))
	fmt.Println("I1")
	for _, line := range lines {
		switch l := line.(type) {
		case PatternLine:
			fmt.Println("I2")
			var lineAttrs lineAttrs
			lineAttrs.attrs = make([]*Attr, 0, len(l.Attrs()))
			fmt.Println("I3")

			resultLine := &patternLine{l.Pattern(), lineAttrs}
			fmt.Println("I4")
			for _, attr := range l.Attrs() {
				macros := mp.macros[attr.K]
				if attr.V == "true" && macros != nil {
					resultLine.attrs = append(
						resultLine.attrs,
						macros...,
					)
				} else if attr.Unspecified && macros != nil {
					for _, m := range macros {
						resultLine.attrs = append(
							resultLine.attrs,
							&Attr{
								K:           m.K,
								Unspecified: true,
							},
						)
					}
				}

				// Git copies through aliases as well as
				// expanding them.
				resultLine.attrs = append(
					resultLine.attrs,
					attr,
				)
			}
			fmt.Println("I5")

			result = append(result, resultLine)
			fmt.Println("I6")

		case MacroLine:
			if readMacros {
				fmt.Println("I7")

				mp.macros[l.Macro()] = l.Attrs()
				fmt.Println("I8")
			}
		}
	}
	return result
}
