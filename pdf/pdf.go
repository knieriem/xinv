package pdf

import (
	"bytes"
	"log"
	"strings"

	"github.com/knieriem/xinv/bl"
	"github.com/knieriem/xinv/sdoc"
	"github.com/signintech/gopdf"
)

type Doc struct {
	pdf  *gopdf.GoPdf
	sDoc *sdoc.Doc
}

type FontSetup struct {
	Setup func(*gopdf.GoPdf) error
}

func NewDoc(fonts *FontSetup) (*Doc, error) {
	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4, Unit: gopdf.UnitCM})
	pdf.AddPage()
	pdf.SetLineWidth(gopdf.PointsToUnits(gopdf.UnitCM, 0.5))

	d := new(Doc)
	d.pdf = pdf

	d.sDoc = sdoc.NewDoc(pdf, gopdf.UnitCM)

	if fonts.Setup != nil {
		err := fonts.Setup(pdf)
		if err != nil {
			return nil, err
		}
	}

	return d, nil
}

func (d *Doc) Bytes() []byte {
	b := new(bytes.Buffer)
	d.pdf.WriteTo(b)
	return b.Bytes()
}

type Action func(sDoc *sdoc.Doc) error

func (d *Doc) Run(act Action) {
	err := act(d.sDoc)
	if err != nil {
		log.Fatal(err)
	}
}

func (d *Doc) AddAddressFieldB(cust, suppl *bl.Party) {
	pdf := d.pdf

	addrX := 2.71
	addrY := 5.08 - gopdf.PointsToUnits(gopdf.UnitCM, 14)
	pdf.SetXY(addrX, addrY)

	// Draw a line below the region where the return address
	// will be written.
	//
	// Note: gopdf.Underline appears too distant from the text,
	// drawing a line gives better results.
	pdf.SetFontSize(8)
	addr := " " + suppl.ReturnAddress
	w, err := pdf.MeasureTextWidth(addr)
	if err == nil {
		pdf.Line(addrX-0.1, addrY+0.1, addrX+w+0.2, addrY+0.1)
	}

	// Write the return address
	pdf.Text(addr)

	d.sDoc.MoveYAbs(sdoc.Cm.Mult(6.27))
	//	d.sDoc.SetReg(sdoc.PointSize, 12)

	sb := new(strings.Builder)
	addLine := func(s string) {
		sb.WriteString(s)
		sb.WriteByte('\n')
	}

	a := cust.Address.GOBLAddress()
	addLine(cust.Name)
	if dept := cust.Contact.Department; dept != "" {
		addLine(dept)
	}
	// FIXME: add cust.Contact.Name if defined
	addLine(a.LineOne())
	if l2 := a.LineTwo(); l2 != "" {
		addLine(l2)
	}
	addLine(string(a.Code) + " " + a.Locality)
	d.sDoc.WriteText(sb.String())
	// d.sDoc.RestoreReg(sdoc.PointSize)
}

func (d *Doc) AddFoldMarksB() {
	pdf := d.pdf
	pdf.SetLineType("")
	pdf.Line(.3, 10.5, .7, 10.5)
	pdf.Line(.3, 14.85, .9, 14.85)
	pdf.Line(.3, 10.5+10.5, .7, 10.5+10.5)
}
