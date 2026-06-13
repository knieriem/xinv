// Package simple provides a layout for simple PDF invoices.
package simple

import (
	"fmt"
	"strconv"
	"strings"

	bc "github.com/bojanz/currency"
	gcal "github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/l10n"
	"golang.org/x/text/language"

	"github.com/knieriem/xinv/bl"
	"github.com/knieriem/xinv/pdf"
	"github.com/knieriem/xinv/pdf/internal/cal"
	"github.com/knieriem/xinv/sdoc"
)

type Invoice struct {
	inv       *bl.Invoice
	formatter *bc.Formatter
	sb        *strings.Builder
	text      *Texts
	country   l10n.Code
}

func NewInvoice(inv *bl.Invoice, texts LocalizedTexts) *Invoice {
	si := new(Invoice)
	si.inv = inv

	country := inv.Customer.Address.Country
	t, ok := texts[l10n.Code(country)]
	if !ok {
		t = &Texts{}
	}
	si.text = t
	si.country = l10n.Code(country)

	si.formatter = bc.NewFormatter(bc.NewLocale(country))
	si.sb = new(strings.Builder)
	return si
}

type LocalizedTexts map[l10n.Code]*Texts

type Texts struct {
	Lang language.Tag

	InvoiceNo   string
	InvoiceDate string
	DateFmt     string

	Email string
	Phone string
	TaxID string

	ColPos   string
	ColItem  string
	ColPrice string

	Total   string
	VAT     string
	Payable string

	ServicePeriod   string
	ServiceDate     string
	DeliveryDate    string
	PaymentTermsTpl string
}

func (si *Invoice) formatAmount(a bl.Amount) string {
	amount, err := bc.NewAmount(a.String(), string(si.inv.Currency))
	if err != nil {
		return err.Error()
	}
	return si.formatter.Format(amount)
}

func (si *Invoice) AddBill() pdf.Action {
	return func(d *sdoc.Doc) error {

		d.VSpace(d.VerticalSpacing())
		d.SetTabs([]sdoc.TabPos{
			{X: sdoc.Cm.Mult(1), Align: sdoc.Center},
			{X: sdoc.Cm.Mult(2.5)},
			{X: sdoc.Cm.Mult(13.5), Align: sdoc.AlignRight},
		})
		d.HLine(sdoc.Cm.Mult(0.3), sdoc.Cm.Mult(14-0.3), sdoc.Cm.Mult(0.2))
		d.WriteText(fmt.Sprintf("\t%s\t%s\t%s   \n", si.text.ColPos, si.text.ColItem, si.text.ColPrice))

		//	d.SetLengthReg(sdoc.PageOffset, sdoc.Centimeters.Mult(2.71+11.5))
		sb := si.sb
		for i := range si.inv.Lines {
			sb.Reset()
			d.VSpace(d.VerticalSpacing())
			line := &si.inv.Lines[i]
			fmt.Fprintf(sb, "\t%d\t%s\t%v\n", i+1, line.Name, si.formatAmount(line.Price))
			desc := strings.TrimRight(line.Description, "\n")
			for desc := range strings.SplitSeq(desc, "\n") {
				fmt.Fprintf(sb, "\t\t%s\n", desc)
			}
			d.WriteText(sb.String())
		}
		d.HLine(sdoc.Cm.Mult(0.3), sdoc.Cm.Mult(14-0.3), 0)
		return nil
	}
}

func (si *Invoice) AddTotals() pdf.Action {
	return func(d *sdoc.Doc) error {
		totals := &si.inv.Totals
		text := si.text
		d.VSpace(d.VerticalSpacing() * 2)
		d.SetTabs([]sdoc.TabPos{
			{X: sdoc.Cm.Mult(10.5), Align: sdoc.AlignRight},
			{X: sdoc.Cm.Mult(13.5), Align: sdoc.AlignRight},
		})
		d.HLine(sdoc.Cm.Mult(6), sdoc.Cm.Mult(8), sdoc.Cm.Mult(0.2))
		d.WriteText(fmt.Sprintf("\t%s\t%v\n", text.Total, si.formatAmount(totals.BaseNet)))

		for i := range totals.Taxes {
			tax := &totals.Taxes[i]
			d.VSpace(d.VerticalSpacing() / 2)
			d.HLine(sdoc.Cm.Mult(6), sdoc.Cm.Mult(8), sdoc.Cm.Mult(0.2))
			d.WriteText(fmt.Sprintf("\t%s %g\u202f%%\t%v\n", text.VAT, tax.Pct.Float64(), si.formatAmount(tax.Tax)))
		}

		d.VSpace(d.VerticalSpacing() / 2)
		d.HLine(sdoc.Cm.Mult(6), sdoc.Cm.Mult(8), sdoc.Cm.Mult(0.2))
		d.WriteText(fmt.Sprintf("\t%s\t%v\n", text.Payable, si.formatAmount(totals.Payable)))
		return nil
	}
}

func (si *Invoice) AddDelivery() pdf.Action {
	return func(d *sdoc.Doc) error {
		sb := si.sb
		sb.Reset()

		d.VSpace(4 * d.VerticalSpacing())
		dlv := &si.inv.Delivery
		if dlv.Date == nil {
			fmt.Fprintf(sb, "%s: %v\n", si.text.ServicePeriod, cal.FormatPeriod(&dlv.Period, si.country))
		} else {
			fmt.Fprintf(sb, "%s: %v\n", si.text.DeliveryDate, cal.FormatDate(dlv.Date, si.country))
		}
		d.WriteText(sb.String())
		return nil
	}
}

func (si *Invoice) AddSupplierInfo() pdf.Action {
	return func(d *sdoc.Doc) error {
		d.SetLengthReg(sdoc.PageOffset, sdoc.Cm.Mult(2.71+11.5))
		d.MoveHAbs(0)
		d.MoveYAbs(sdoc.Cm.Mult(1.6 + .5))
		d.SetTabs([]sdoc.TabPos{
			{X: sdoc.Cm.Mult(1.4)},
		})

		sb := si.sb
		sb.Reset()
		suppl := si.inv.Supplier
		a := suppl.Address.GOBLAddress()

		addLine := func(sb *strings.Builder, s string) {
			sb.WriteString(s)
			sb.WriteByte('\n')
		}
		addLine(sb, suppl.Name)
		if tn := suppl.TradeName; tn != "" {
			addLine(sb, suppl.TradeName)
		}
		addLine(sb, a.LineOne())
		if l2 := a.LineTwo(); l2 != "" {
			addLine(sb, a.LineTwo())
		}
		addLine(sb, a.Code.String()+" "+a.Locality)

		text := si.text
		sb.WriteByte('\n')
		fmt.Fprintf(sb, "%s\t%s\n", text.Email, suppl.Email)
		fmt.Fprintf(sb, "%s\t%s\n", text.Phone, suppl.Phone)
		sb.WriteByte('\n')
		fmt.Fprintf(sb, "%s %s\n", text.TaxID, suppl.TaxID)

		d.WriteText(sb.String())

		d.RestoreReg(sdoc.PageOffset)

		d.SetTabs([]sdoc.TabPos{
			{X: sdoc.Cm.Mult(3.5)},
			{X: sdoc.Cm.Mult(11.5), Align: sdoc.AlignRight},
		})
		return nil
	}
}

func (si *Invoice) AddReference() pdf.Action {
	return func(d *sdoc.Doc) error {
		d.MoveYAbs(sdoc.Cm.Mult(10))
		d.MoveHAbs(0)
		sb := si.sb
		sb.Reset()
		text := si.text
		fmt.Fprintf(sb, "%s %s\t%s: %v\n", text.InvoiceNo, si.inv.Code, text.InvoiceDate, si.inv.IssueTime.Format(text.DateFmt))
		if p := si.inv.Ordering.ProjectName; p != "" {
			fmt.Fprintf(sb, "\nProjekt: %s\n", p) // FIXME
		}
		d.WriteText(sb.String())
		//	d.RestoreReg(sdoc.PageOffset)
		return nil
	}
}

func (si *Invoice) AddPaymentTerms() pdf.Action {
	return func(d *sdoc.Doc) error {
		sb := si.sb
		sb.Reset()
		terms := &si.inv.Supplier.Payments.Terms
		if terms.DueDays != 0 {
			issueDate := gcal.MakeDate(si.inv.IssueTime.AddDate(0, 0, terms.DueDays).Date())
			dateStr := cal.FormatDate(&issueDate, si.country)
			r := strings.NewReplacer("{{.NumDays}}", strconv.Itoa(terms.DueDays), "{{.DueDate}}", dateStr)
			fmt.Fprintf(sb, "%s", r.Replace(si.text.PaymentTermsTpl))
		}
		d.VSpace(d.VerticalSpacing() / 2)
		d.WriteText(sb.String())
		return nil
	}
}

func (si *Invoice) AddPaymentInstr() pdf.Action {
	return func(d *sdoc.Doc) error {
		d.MoveYAbs(sdoc.Cm.Mult(29.7 - 3))
		d.MoveHAbs(0)

		sb := si.sb
		sb.Reset()
		instr := &si.inv.Supplier.Payments.Instr
		if instr.SEPA != nil {
			fmt.Fprintln(sb, instr.SEPA.Bank)
			fmt.Fprintf(sb, "IBAN %s\n", instr.SEPA.IBAN)
		}
		d.GoPDF().SetTextColor(100, 100, 100)
		d.WriteText(sb.String())
		d.GoPDF().SetTextColor(0, 0, 0)
		return nil
	}
}
