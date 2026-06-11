package conflict

import (
	"bufio"
	"io"
	"strings"
)

type Block struct {
	Ours       string
	Theirs     string
	OursLabel  string
	TheirsLabel string
}

func (b Block) ReEmit() string {
	var out strings.Builder
	out.WriteString("<<<<<<< " + b.OursLabel)
	if b.Ours != "" {
		out.WriteString("\n" + b.Ours)
	}
	out.WriteString("\n=======")
	if b.Theirs != "" {
		out.WriteString("\n" + b.Theirs)
	}
	out.WriteString("\n>>>>>>> " + b.TheirsLabel)
	return out.String()
}

type File struct {
	Sections []Section
}

type Section struct {
	Text     string
	Conflict *Block
}

func Parse(r io.Reader) (*File, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, bufio.MaxScanTokenSize), 1024*1024)
	var f File
	var current strings.Builder
	var ours, theirs strings.Builder
	var oursLabel, theirsLabel string
	inOurs, inTheirs := false, false
	currentHasLines := false

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "<<<<<<< "):
			f.Sections = append(f.Sections, Section{Text: current.String()})
			current.Reset()
			currentHasLines = false
			ours.Reset()
			theirs.Reset()
			oursLabel = strings.TrimPrefix(line, "<<<<<<< ")
			inOurs = true
			inTheirs = false

		case line == "=======" && inOurs:
			inOurs = false
			inTheirs = true

		case strings.HasPrefix(line, ">>>>>>> ") && inTheirs:
			theirsLabel = strings.TrimPrefix(line, ">>>>>>> ")
			f.Sections = append(f.Sections, Section{
				Conflict: &Block{
					Ours:        ours.String(),
					Theirs:      theirs.String(),
					OursLabel:   oursLabel,
					TheirsLabel: theirsLabel,
				},
			})
			inOurs = false
			inTheirs = false

		case inOurs:
			if ours.Len() > 0 {
				ours.WriteByte('\n')
			}
			ours.WriteString(line)

		case inTheirs:
			if theirs.Len() > 0 {
				theirs.WriteByte('\n')
			}
			theirs.WriteString(line)

		default:
			if currentHasLines {
				current.WriteByte('\n')
			}
			current.WriteString(line)
			currentHasLines = true
		}
	}

	if currentHasLines {
		f.Sections = append(f.Sections, Section{Text: current.String()})
	}

	return &f, scanner.Err()
}

func (f *File) HasConflicts() bool {
	for _, s := range f.Sections {
		if s.Conflict != nil {
			return true
		}
	}
	return false
}

func (f *File) Conflicts() []Block {
	var blocks []Block
	for _, s := range f.Sections {
		if s.Conflict != nil {
			blocks = append(blocks, *s.Conflict)
		}
	}
	return blocks
}

func (f *File) Render(resolver func(Block) string) string {
	var out strings.Builder
	first := true
	for _, s := range f.Sections {
		if s.Conflict != nil {
			resolved := resolver(*s.Conflict)
			if out.Len() > 0 {
				out.WriteByte('\n')
			}
			out.WriteString(resolved)
			first = false
		} else if s.Text != "" {
			if !first {
				out.WriteByte('\n')
			}
			out.WriteString(s.Text)
			first = false
		}
	}
	out.WriteByte('\n')
	return out.String()
}
