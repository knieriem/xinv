// Package sdoc provides some troff inspired utilites for PDF
// creation using gopdf.
package sdoc

import (
	"strings"

	"github.com/signintech/gopdf"
)

type Length uint32

func (l Length) Mult(f float64) Length {
	return Length(f * float64(l))
}

func (l Length) intVal() int {
	return int(l)
}

const (
	Pt Length = 64
	Cm Length = 64 * 7200 / 254
)

func (l Length) toPDFUnit(unit int) float64 {
	if unit == gopdf.UnitCM {
		return float64(l) / float64(Cm)
	}
	return 0
}

func (d *Doc) lengthFromPDFUnit(l float64) Length {
	return Length(l * float64(Cm))
}

type Reg int

const (
	VerticalSpacing Reg = iota
	PageOffset
	PointSize
	numRegs
)

type Doc struct {
	pdf  *gopdf.GoPdf
	unit int

	regs     [numRegs]int
	prevRegs [numRegs]int

	x, y Length
	tabs []TabPos
}

func (d *Doc) SetReg(r Reg, v int) {
	d.prevRegs[r] = d.regs[r]
	d.regs[r] = v
}

func (d *Doc) SetLengthReg(r Reg, l Length) {
	d.prevRegs[r] = d.regs[r]
	d.regs[r] = l.intVal()
}

func (d *Doc) RestoreReg(r Reg) {
	d.regs[r] = d.prevRegs[r]
}

func NewDoc(pdf *gopdf.GoPdf, pdfUnit int) *Doc {
	d := Doc{
		pdf:  pdf,
		unit: pdfUnit,
	}
	return &d
}

func (d *Doc) GoPDF() *gopdf.GoPdf {
	return d.pdf
}

func (d *Doc) TextWidth(s string) Length {
	w, err := d.pdf.MeasureTextWidth(s)
	if err != nil {
		return 0
	}
	return d.lengthFromPDFUnit(w)
}

func (d *Doc) MoveYAbs(y Length) {
	d.y = y
}
func (d *Doc) MoveHAbs(x Length) {
	d.x = x
}

func (d *Doc) VSpace(y Length) {
	d.y += y
}

func (d *Doc) VerticalSpacing() Length {
	return Length(d.regs[VerticalSpacing])
}

func (d *Doc) WriteText(text string) {
	pdf := d.pdf
	pdf.SetFontSize(float64(d.regs[PointSize]))

	x := d.x
	withinLine := !strings.HasSuffix(text, "\n")
	text = strings.TrimRight(text, "\n")
	f := strings.Split(text, "\n")
	for i, line := range f {
		if line != "" {
			d.writeLineTabbed(line, x, d.tabs)
		}
		if i == len(f)-1 && withinLine {
			x = d.lengthFromPDFUnit(pdf.GetX()) - Length(d.regs[PageOffset])
			break
		}
		d.y += Length(d.regs[VerticalSpacing])
		x = 0
	}
	d.x = x
}

func (d *Doc) writeLineTabbed(line string, x0 Length, tabs []TabPos) {
	pdf := d.pdf

	t := newTabber(tabs)
	for t.next() {
		s := line
		i := strings.IndexByte(s, '\t')
		eol := i == -1
		if !eol {
			s = s[:i]
			line = line[i+1:]
		} else {
			line = line[:0]
		}

		x := t.x
		div := Length(1)
		if t.cur != -1 {
			tab := &t.tabs[t.cur]
			switch tab.Align {
			case AlignLeft:
			case Center:
				div = 2
				fallthrough
			case AlignRight:
				w := d.TextWidth(s)
				x -= w / div

			}
		} else {
			x = x0
		}
		pdf.SetXY((x + Length(d.regs[PageOffset])).toPDFUnit(d.unit), Length(d.y).toPDFUnit(d.unit))
		pdf.Text(s)

		if eol {
			break
		}
	}
}

type TabPos struct {
	X     Length
	Incr  bool
	Align Alignment
}

type Alignment int

const (
	AlignLeft Alignment = iota
	AlignRight
	Center
)

type tabber struct {
	tabs []TabPos
	cur  int
	x    Length
}

func newTabber(tabs []TabPos) *tabber {
	return &tabber{cur: -2, tabs: tabs}
}

func (t *tabber) next() bool {
	t.cur++
	if t.cur >= len(t.tabs) {
		return false
	}

	if t.cur < 0 {
		return true
	}
	tab := &t.tabs[t.cur]
	if tab.Incr {
		t.x += tab.X
	} else {
		t.x = tab.X
	}
	return true
}

func (d *Doc) SetTabs(tabs []TabPos) {
	d.tabs = tabs
}

func (d *Doc) HLine(x, l, yOff Length) {
	x = Length(d.regs[PageOffset]) + x
	x0 := x.toPDFUnit(d.unit)
	x1 := (x + l).toPDFUnit(d.unit)
	y := (d.y + yOff).toPDFUnit(d.unit)
	d.pdf.Line(x0, y, x1, y)
}
