package gitattr

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
func (mp *MacroProcessor) ProcessLines(lines []*Line, readMacros bool) []*Line {
	result := make([]*Line, 0, len(lines))
	for _, line := range lines {
		if line.Pattern != nil {
			resultLine := Line{
				Pattern: line.Pattern,
				Attrs:   make([]*Attr, 0, len(line.Attrs)),
			}
			for _, attr := range line.Attrs {
				macros := mp.macros[attr.K]
				if attr.V == "true" && macros != nil {
					resultLine.Attrs = append(
						resultLine.Attrs,
						macros...,
					)
				}

				// Git copies through aliases as well as
				// expanding them.
				resultLine.Attrs = append(
					resultLine.Attrs,
					attr,
				)
			}
			result = append(result, &resultLine)
		} else if readMacros {
			mp.macros[line.Macro] = line.Attrs
		}
	}
	return result
}
