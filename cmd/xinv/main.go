// cmd/xinv creates a Factur-X/ZUGFeRD PDF document from
// an invoice text file.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/knieriem/xinv"
	"github.com/knieriem/xinv/bl"
	"github.com/knieriem/xinv/cmd/xinv/internal/ti"
	"github.com/knieriem/xinv/facturx"
	"github.com/knieriem/xinv/pdf"
	"github.com/knieriem/xinv/pdf/fonts/arimo"
	"github.com/knieriem/xinv/pdf/simple"
	"github.com/knieriem/xinv/sdoc"
)

var (
	confDir   = flag.String("C", "config", "configuration directory")
	outputPDF = flag.String("o", "", "output PDF filename")

	issueTime time.Time
)

type config struct {
	Customer       []bl.Party
	Supplier       []bl.Party
	LocalizedTexts simple.LocalizedTexts
}

type doc struct {
	Invoice bl.InvoiceSrc
}

func main() {
	flag.TextVar(&issueTime, "date", time.Time{}, "issue date")
	flag.Parse()

	if *outputPDF == "" {
		errExit(errors.New("output filename unset"))
	}

	conf, err := parseConfig()
	if err != nil {
		errExit(err)
	}

	var invDoc doc
	err = ti.ParseFile(flag.Arg(0), &invDoc)
	if err != nil {
		errExit(err)
	}

	if issueTime.IsZero() {
		issueTime = time.Now()
	}

	err = createXInvoice(*outputPDF, &invDoc, issueTime, conf)
	if err != nil {
		errExit(err)
	}
}

func createXInvoice(outFilename string, src *doc, issueTime time.Time, c *config) error {

	inst := xinv.NewInstance(c.Customer, c.Supplier)
	inv, err := inst.MakeInvoice(&src.Invoice, issueTime)
	if err != nil {
		return err
	}

	fonts := pdf.FontSetup{
		Setup: arimo.Setup,
	}

	doc, err := pdf.NewDoc(&fonts)
	if err != nil {
		return err
	}

	doc.Run(func(d *sdoc.Doc) error {
		d.SetLengthReg(sdoc.PageOffset, sdoc.Cm.Mult(2.71))
		d.SetLengthReg(sdoc.VerticalSpacing, 16*sdoc.Pt)
		d.SetReg(sdoc.PointSize, 12)
		d.SetTabs([]sdoc.TabPos{{X: sdoc.Cm.Mult(11.5)}})
		return nil
	})

	doc.AddFoldMarksB()
	doc.AddAddressFieldB(inv.Customer, inv.Supplier)

	si := simple.NewInvoice(inv, c.LocalizedTexts)
	doc.Run(si.AddPaymentInstr())
	doc.Run(si.AddReference())
	doc.Run(si.AddBill())
	doc.Run(si.AddTotals())
	doc.Run(si.AddDelivery())
	doc.Run(si.AddPaymentTerms())
	doc.Run(si.AddSupplierInfo())

	pdfBytes := doc.Bytes()

	fxi := facturx.InvoiceData{
		Code:         inv.Code,
		IssueTime:    issueTime,
		SupplierName: inv.Supplier.Name,
		ZUGFeRDV2XML: inv.ZUGFeRDV2XML,
	}

	out := new(bytes.Buffer)
	err = facturx.WriteDoc(out, pdfBytes, &fxi)
	if err != nil {
		return err
	}

	f, err := os.Create(outFilename)
	if err != nil {
		return err
	}
	defer f.Close()
	out.WriteTo(f)
	return nil
}

func parseConfig() (*config, error) {

	fsys := os.DirFS(*confDir)

	var c config
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		f, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		return ti.Parse(f, &c)
	})
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func errExit(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
