package xinv_test

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"

	"github.com/knieriem/xinv"
	"github.com/knieriem/xinv/bl"
	"github.com/knieriem/xinv/facturx"
	"github.com/knieriem/xinv/pdf"
	"github.com/knieriem/xinv/pdf/fonts/arimo"
	"github.com/knieriem/xinv/pdf/simple"
	"github.com/knieriem/xinv/sdoc"
)

type config struct {
	Customer       []bl.Party
	Supplier       []bl.Party
	LocalizedTexts simple.LocalizedTexts
}

type doc struct {
	Invoice bl.InvoiceSrc
}

type test struct {
	id string
	c  *config

	issueTime time.Time
}

func TestInvoice(t *testing.T) {
	var conf config

	err := parseJSON(filepath.Join("testdata", "config.json"), &conf)
	if err != nil {
		t.Fatal(err)
	}

	test := test{
		id:        "1",
		c:         &conf,
		issueTime: time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC),
	}
	t.Run("testInvoiceParts", test.testInvoiceParts)
}

func (test *test) filePath(name string) string {
	return filepath.Join("testdata", test.id, name)
}

func (test *test) testInvoiceParts(t *testing.T) {
	var invDoc doc
	err := parseJSON(test.filePath("invoice.json"), &invDoc)
	if err != nil {
		t.Fatal(err)
	}
	inst := xinv.NewInstance(test.c.Customer, test.c.Supplier)
	inv, err := inst.MakeInvoice(&invDoc.Invoice, test.issueTime)
	if err != nil {
		t.Fatal(err)
	}

	haveJSON := invoiceJSONNoUUIDs(inv.GOBLData(), t)
	wantJSON := invoiceJSONNoUUIDs(goblEnvFromFile(test.filePath("gobl.json"), t), t)
	if !bytes.Equal(wantJSON, haveJSON) {
		t.Fatal("GOBL documents not equal")
	}

	fonts := pdf.FontSetup{
		Setup: arimo.Setup,
	}

	doc, err := pdf.NewDoc(&fonts)
	if err != nil {
		t.Fatal(err)
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

	si := simple.NewInvoice(inv, test.c.LocalizedTexts)
	doc.Run(si.AddPaymentInstr())
	doc.Run(si.AddReference())
	doc.Run(si.AddBill())
	doc.Run(si.AddTotals())
	doc.Run(si.AddDelivery())
	doc.Run(si.AddPaymentTerms())
	doc.Run(si.AddSupplierInfo())

	plainPDF := doc.Bytes()
	test.compareData("plain.pdf", plainPDF, t)

	fxi := facturx.InvoiceData{
		Code:         inv.Code,
		IssueTime:    test.issueTime,
		SupplierName: inv.Supplier.Name,
		ZUGFeRDV2XML: inv.ZUGFeRDV2XML,
	}

	out := new(bytes.Buffer)
	err = facturx.WriteDoc(out, plainPDF, &fxi)
	if err != nil {
		t.Fatal(err)
	}

	test.compareData("invoice.pdf", out.Bytes(), t)
}

func goblEnvFromFile(filename string, t *testing.T) *gobl.Envelope {
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	ob, err := gobl.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	env, ok := ob.(*gobl.Envelope)
	if !ok {
		t.Fatal("GOBL object does not contain envelope")
	}
	return env
}

func invoiceJSONNoUUIDs(env *gobl.Envelope, t *testing.T) []byte {
	inv, ok := env.Extract().(*bill.Invoice)
	if !ok {
		t.FailNow()
	}
	inv.UUID = ""
	b, err := json.Marshal(inv)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (test *test) compareData(wantFilename string, have []byte, t *testing.T) {
	want, err := os.ReadFile(test.filePath(wantFilename))
	if err != nil {
		t.Fatal(err)
	}

	os.WriteFile(",,have."+wantFilename, have, 0644)
	if !bytes.Equal(have, want) {
		t.Errorf("%q does not match", wantFilename)
	}
}

func parseJSON(filename string, data any) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	return d.Decode(data)
}
