// Package facturx creates a Factur-X document from
// an existing PDF and an XML invoice.
package facturx

import (
	"bytes"
	"io"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/knieriem/xinv/facturx/internal/fx"
	"github.com/knieriem/xinv/facturx/internal/pdfa"
)

type InvoiceData struct {
	Code         string
	IssueTime    time.Time
	SupplierName string
	ZUGFeRDV2XML []byte
}

func WriteDoc(w io.Writer, pdfData []byte, inv *InvoiceData) error {
	inFile := bytes.NewReader(pdfData)
	ctx, err := api.ReadAndValidate(inFile, model.NewDefaultConfiguration())
	if err != nil {
		return err
	}

	invoiceData := fx.InvoiceData{
		Date:    inv.IssueTime,
		XMLData: inv.ZUGFeRDV2XML,
		Spec:    fx.ZUGFeRDv21Source,
	}

	err = fx.EmbedZUGFeRDAttachment(ctx, &invoiceData)
	if err != nil {
		return err
	}

	xmp := &fx.XMPData{
		Invoice: fx.XMPInvoiceData{
			Number:      inv.Code,
			Date:        inv.IssueTime,
			Supplier:    inv.SupplierName,
			ProfileName: fx.EN16931,
		},

		CreatorTool: "xinv",
		PDFProducer: "pdfcpu " + model.VersionStr,
	}

	err = pdfa.SetOutputIntent(ctx, &pdfa.OutputIntent_sRGB)
	if err != nil {
		return err
	}

	err = fx.SetMetaInfo(ctx, xmp)
	if err != nil {
		return err
	}
	api.OptimizeContext(ctx)

	ctx.Write.CreationDateOverride = inv.IssueTime
	return api.WriteContext(ctx, w)
}
