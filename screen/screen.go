package screen

import (
	termbox "github.com/nsf/termbox-go"
)

type Line struct {
	Line  string
	Color termbox.Attribute
}

type Screen struct {
	Lines      []Line
	Editing    bool
	CurPath    string
	CursorPosX int
	EditLine   []rune
}

func NewScreen(curPath string) *Screen {
	return &Screen{
		Lines:      []Line{},
		Editing:    true,
		CurPath:    curPath,
		CursorPosX: 0,
	}
}

func (s *Screen) Redraw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	_, height := termbox.Size()

	y := height - 1

	if s.Editing {
		x := 0
		for _, c := range []rune(s.CurPath + " > ") {
			termbox.SetCell(x, y, c, termbox.ColorYellow, termbox.ColorDefault)
			x++
		}

		startLineIdx := x
		setCursorCell := false
		for _, c := range s.EditLine {
			if !setCursorCell && s.CursorPosX+startLineIdx == x {
				setCursorCell = true
				termbox.SetCell(x, y, c, termbox.ColorBlack, termbox.ColorWhite)
			} else {
				termbox.SetCell(x, y, c, termbox.ColorWhite, termbox.ColorDefault)
			}
			x++
		}

		if !setCursorCell {
			termbox.SetCell(x, y, ' ', termbox.ColorBlack, termbox.ColorWhite)
		}

		y--
	}

	for i := len(s.Lines) - 1; i >= 0; i-- {
		if y < 0 {
			break
		}

		chars := []rune(s.Lines[i].Line)
		for x := 0; x < len(chars); x++ {
			termbox.SetCell(x, y, chars[x], s.Lines[i].Color, termbox.ColorDefault)
		}

		y--
	}

	termbox.Flush()
}

func (s *Screen) MoveCursorLeft(redraw bool) {
	if !s.Editing || s.CursorPosX == 0 {
		return
	}
	s.CursorPosX--
	if redraw {
		s.Redraw()
	}
}

func (s *Screen) MoveCursorRight(redraw bool) {
	if !s.Editing || s.CursorPosX == len(s.EditLine) {
		return
	}

	s.CursorPosX++
	if redraw {
		s.Redraw()
	}
}

func (s *Screen) AppendLines(redraw bool, color termbox.Attribute, lines ...string) {
	for _, line := range lines {
		s.Lines = append(s.Lines, Line{Line: line, Color: color})
	}
	if redraw {
		s.Redraw()
	}
}

func (s *Screen) AppendAtCursor(redraw bool, runes ...rune) {
	if !s.Editing {
		return
	}
	if s.CursorPosX >= len(s.EditLine) {
		s.EditLine = append(s.EditLine, runes...)
		s.CursorPosX = len(s.EditLine)
	} else {
		newLine := s.EditLine[:s.CursorPosX]
		newLine = append(newLine, runes...)
		s.EditLine = append(newLine, s.EditLine[s.CursorPosX:]...)
		s.CursorPosX += len(runes)
	}
	if redraw {
		s.Redraw()
	}
}

func (s *Screen) BackspaceAtCursor(redraw bool) {
	if !s.Editing || s.CursorPosX == 0 {
		return
	}
	if s.CursorPosX >= len(s.EditLine) {
		s.EditLine = s.EditLine[:len(s.EditLine)-1]
	} else {
		newLine := s.EditLine[:s.CursorPosX-1]
		s.EditLine = append(newLine, s.EditLine[s.CursorPosX:]...)
	}
	s.CursorPosX--
	if redraw {
		s.Redraw()
	}
}

func (s *Screen) SetEditLine(redraw bool, editLine []rune) {
	s.EditLine = editLine
	s.CursorPosX = len(s.EditLine)
	if redraw {
		s.Redraw()
	}
}
