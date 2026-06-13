package xinv

import (
	"fmt"
	"time"

	"github.com/knieriem/xinv/bl"
)

type Instance struct {
	Customers map[string]*bl.Party
	Suppliers map[string]*bl.Party
}

func NewInstance(customers, suppliers []bl.Party) *Instance {
	inst := new(Instance)

	inst.Customers = partyMap(customers)
	inst.Suppliers = partyMap(suppliers)
	return inst
}

func partyMap(list []bl.Party) map[string]*bl.Party {
	m := make(map[string]*bl.Party, len(list))
	for i := range list {
		p := &list[i]
		m[p.ID] = p
	}
	return m
}

func (inst *Instance) MakeInvoice(src *bl.InvoiceSrc, issueTime time.Time) (*bl.Invoice, error) {

	cust, ok := inst.Customers[src.CustomerID]
	if !ok {
		return nil, fmt.Errorf("unknown customer %q", src.CustomerID)
	}

	suppl, ok := inst.Suppliers[src.SupplierID]
	if !ok {
		return nil, fmt.Errorf("unknown supplier %q", src.SupplierID)
	}

	inv := new(bl.Invoice)
	inv.IssueTime = issueTime
	inv.Customer = cust
	inv.Supplier = suppl
	inv.InvoiceSrc = src
	err := inv.Calculate()
	if err != nil {
		return nil, err
	}
	return inv, nil
}
