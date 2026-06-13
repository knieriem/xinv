package fx

import (
	_ "embed"
	"strings"
	"text/template"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

type Profile string

const (
	BASIC    Profile = "BASIC"
	EN16931  Profile = "EN 16931"
	EXTENDED Profile = "EXTENDED"
)

// XMPData defines PDF/A meta data
type XMPData struct {
	Invoice     XMPInvoiceData
	CreatorTool string
	PDFProducer string
}

type XMPInvoiceData struct {
	Number      string
	Date        time.Time
	Supplier    string
	ProfileName Profile
}

//go:embed "factur-x.xmp.tpl"
var xmpTplStr string

var xmpTpl = template.Must(template.New("").Parse(xmpTplStr))

func SetMetaInfo(ctx *model.Context, data *XMPData) error {
	xRefTable := xRefTable{ctx.XRefTable}
	rootDict, err := xRefTable.Catalog()
	if err != nil {
		return err
	}

	sb := new(strings.Builder)
	err = xmpTpl.Execute(sb, data)
	if err != nil {
		return err
	}
	s := sb.String()

	md, err := xRefTable.newMetadataStreamDict(strings.NewReader(s))
	if err != nil {
		return err
	}
	rootDict.Update("Metadata", *md)
	return nil
}
