// Package bl provides a simplified interface to GOBL.
package bl

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/invopop/gobl"
	cii "github.com/invopop/gobl.cii"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/pay"
	"github.com/invopop/gobl/tax"
)

type InvoiceSrc struct {
	Code       string
	CustomerID string
	SupplierID string
	Ordering   Ordering

	Delivery Delivery

	Lines []BillLine
}

type Ordering struct {
	Code        string // FIXME N/A if empty
	ProjectID   string // FIXME
	ProjectName string
}

type Delivery struct {
	Date   *Date
	Period Period
}

type BillLine struct {
	Quantity    int
	Name        string
	Description string
	Price       Amount
}

type Invoice struct {
	*InvoiceSrc
	IssueTime time.Time
	Supplier  *Party
	Customer  *Party

	Currency Currency
	Totals   Totals

	ZUGFeRDV2XML []byte
}

type Party struct {
	ID            string
	Name          string
	TradeName     string
	Address       Address
	ReturnAddress string
	Email         string
	Phone         string
	Contact       Person
	MachineEmail  string
	TaxID         string
	Payments      Payments
}

type Address struct {
	Number      string
	Street      string
	StreetExtra string
	Locality    string
	Code        string
	Country     string
}

func (a *Address) GOBLAddress() *org.Address {
	return &org.Address{
		Number:      a.Number,
		Street:      a.Street,
		StreetExtra: a.StreetExtra,

		Locality: a.Locality,
		Code:     cbc.Code(a.Code),
		Country:  l10n.ISOCountryCode(a.Country),
	}
}

type Person struct {
	Name       Name
	Department string
	Email      string
}

type Name = org.Name

type Date = cal.Date

type Period = cal.Period

type Amount = num.Amount

type Currency = currency.Code

type Totals struct {
	BaseNet Amount
	Taxes   []TaxAmount
	Payable Amount
}

type TaxAmount struct {
	Pct Amount
	Tax Amount
}

type Payments struct {
	Terms struct {
		DueDays int
	}
	Instr struct {
		SEPA *CreditTransfer
	}
}

type CreditTransfer struct {
	Bank string
	IBAN IBAN
}

type IBAN string

func (iban IBAN) narrow() string {
	return strings.ReplaceAll(string(iban), " ", "")
}

func (p *Party) deriveCountryCode() (country l10n.ISOCountryCode, tax l10n.TaxCountryCode, err error) {
	if len(p.TaxID) < 2 {
		return "", "", errors.New("taxID too short")
	}
	switch p.TaxID[:2] {
	case "DE":
		tax = l10n.TaxCountryCode(l10n.DE)
	default:
		return "", "", errors.New("cannot derive tax country from tax id")
	}

	cc := l10n.Code(p.Address.Country)
	country = l10n.ISOCountryCode(cc)
	if country == "" {
		country = l10n.ISOCountryCode(tax)
		p.Address.Country = string(tax)
	}
	return country, tax, nil
}

func (p *Party) goblParty() (*org.Party, error) {
	_, taxCountry, err := p.deriveCountryCode()
	if err != nil {
		return nil, fmt.Errorf("cannot derive country code from supplier: %w", err)
	}

	pp := new(org.Party)
	pp.Name = p.Name
	pp.TaxID = &tax.Identity{
		Country: taxCountry,
		Code:    "DE123456789",
	}
	pp.Addresses = []*org.Address{
		p.Address.GOBLAddress(),
	}
	email := p.Email

	machineEmail := p.MachineEmail
	if machineEmail == "" {
		machineEmail = email
	}
	pp.Inboxes = []*org.Inbox{{Email: machineEmail}}
	pp.Emails = []*org.Email{{Address: email}}
	if p.Phone != "" {
		pp.Telephones = []*org.Telephone{{Number: p.Phone}}
	}
	contactEmail := p.Contact.Email
	if contactEmail == "" {
		contactEmail = email
	}
	pp.People = []*org.Person{
		{
			Name:   &p.Contact.Name,
			Role:   p.Contact.Department,
			Emails: []*org.Email{{Address: contactEmail}},
		},
	}
	return pp, nil
}

func (inv *Invoice) Calculate() error {

	customer, err := inv.Customer.goblParty()
	if err != nil {
		return err
	}
	supplier, err := inv.Supplier.goblParty()
	if err != nil {
		return err
	}

	issueDate := cal.MakeDate(inv.IssueTime.Date())

	bi := &bill.Invoice{
		Code:      cbc.Code(inv.Code),
		IssueDate: issueDate,
		Customer:  customer,
		Supplier:  supplier,
		Addons:    tax.WithAddons(cbc.Key("de-zugferd-v2")),
	}

	o := new(bill.Ordering)
	oUsed := false
	if inv.Ordering.Code != "" {
		o.Code = cbc.Code(inv.Ordering.Code)
		oUsed = true
	}
	if pID, pName := inv.Ordering.ProjectID, inv.Ordering.ProjectName; pID != "" || pName != "" {
		if pID == "" {
			pID = pName
		} else if pName == "" {
			pName = pID
		}
		o.Projects = []*org.DocumentRef{
			{
				Code:        cbc.Code(pID),
				Description: pName,
			},
		}
		oUsed = true
	}
	if oUsed {
		bi.Ordering = o
	}

	dy := inv.Delivery
	if dy.Date != nil {
		bi.Delivery = &bill.DeliveryDetails{Date: dy.Date}
	} else {
		per := &dy.Period
		bi.Delivery = &bill.DeliveryDetails{Date: &per.End, Period: &cal.Period{Start: per.Start, End: per.End}}
	}

	bi.Lines = make([]*bill.Line, len(inv.Lines))
	for i := range inv.Lines {
		src := &inv.Lines[i]
		line := new(bill.Line)
		q := src.Quantity
		if q == 0 {
			q = 1
		}
		line.Quantity = num.MakeAmount(int64(q), 0)
		line.Item = &org.Item{
			Name:        src.Name,
			Description: src.Description,
			Price:       &src.Price,
		}
		line.Taxes = []*tax.Combo{
			{Category: tax.CategoryVAT, Rate: "standard"},
		}
		bi.Lines[i] = line
	}

	bi.Payment = new(bill.PaymentDetails)

	pmt := inv.Supplier.Payments
	if pmt.Terms.DueDays != 0 {
		dueDate := cal.Date{Date: issueDate.AddDays(30)}
		bi.Payment.Terms = &pay.Terms{
			DueDates: []*pay.DueDate{{Date: &dueDate}},
		}
	}
	if ct := pmt.Instr.SEPA; ct != nil {
		bi.Payment.Instructions = &pay.Instructions{
			Key: "credit-transfer",
			CreditTransfer: []*pay.CreditTransfer{
				{
					Name: ct.Bank,
					IBAN: ct.IBAN.narrow(),
				},
			},
		}
	}

	env := gobl.NewEnvelope()
	if err := env.Insert(bi); err != nil {
		return err
	}

	// run standard validations, tax logic compilation, and sub-total sums
	if err := env.Calculate(); err != nil {
		return err
	}
	inv.Totals.BaseNet = bi.Totals.Total
	inv.Totals.Payable = bi.Totals.Payable
	for _, cat := range bi.Totals.Taxes.Categories {
		if cat.Code != tax.CategoryVAT {
			return fmt.Errorf("unable to handle non-VAT tax: %q", cat.Code)
		}
		tx := make([]TaxAmount, len(cat.Rates))
		for i, rate := range cat.Rates {
			tx[i].Tax = rate.Amount
			tx[i].Pct = rate.Percent.Amount()
		}
		inv.Totals.Taxes = tx
	}
	inv.Currency = bi.Currency
	bi.Payment.Terms.DueDates[0].Amount = bi.Totals.Payable

	doc, err := cii.ConvertInvoice(env, cii.WithContext(cii.ContextZUGFeRDV2))
	if err != nil {
		return err
	}

	xmlData, err := doc.Bytes()
	if err != nil {
		return err
	}
	inv.ZUGFeRDV2XML = xmlData

	return nil
}
