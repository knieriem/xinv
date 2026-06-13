package fx

import (
	"bytes"
	"io"
	"time"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type InvoiceData struct {
	XMLData []byte
	Date    time.Time
	Spec    *VersionSpec
}

type VersionSpec struct {
	EmbeddedFilename string
	EmbeddedFileDesc string
	AFRelationship   string
}

var ZUGFeRDv20 = &VersionSpec{
	EmbeddedFilename: "zugferd-invoice.xml",
	EmbeddedFileDesc: "ZUGFeRD XML invoice",
}

var ZUGFeRDv21 = &VersionSpec{
	EmbeddedFilename: "factur-x.xml",
	EmbeddedFileDesc: "ZUGFeRD/Factur-X XML invoice",
	AFRelationship:   "Alternative",
}

var FacturX = &VersionSpec{
	EmbeddedFilename: "factur-x.xml",
	EmbeddedFileDesc: "Factur-X XML invoice",
}

var ZUGFeRDv21Source = &VersionSpec{
	EmbeddedFilename: "factur-x.xml",
	EmbeddedFileDesc: "Factur-X XML invoice",
	AFRelationship:   "Source",
}

type xRefTable struct {
	*model.XRefTable
}

func EmbedZUGFeRDAttachment(ctx *model.Context, data *InvoiceData) error {
	xRefTable := &xRefTable{ctx.XRefTable}
	if err := xRefTable.LocateNameTree("EmbeddedFiles", true); err != nil {
		return err
	}

	if false {
		// Ensure a Collection entry in the catalog.
		if err := xRefTable.EnsureCollection(); err != nil {
			return err
		}
	}

	a := model.Attachment{
		ID:       data.Spec.EmbeddedFilename,
		FileName: data.Spec.EmbeddedFilename,
		Desc:     data.Spec.EmbeddedFileDesc,
		Reader:   bytes.NewReader(data.XMLData),
		ModTime:  &data.Date,
	}

	modTime := data.Date
	sd, err := xRefTable.newEmbeddedStreamDict(a, modTime)
	if err != nil {
		return err
	}

	d, err := xRefTable.newFileSpecDict(a.ID, a.ID, a.Desc, data.Spec.AFRelationship, *sd)
	if err != nil {
		return err
	}

	ir, err := xRefTable.IndRefForNewObject(d)
	if err != nil {
		return err
	}

	m := model.NameMap{a.ID: []types.Dict{d}}

	err = xRefTable.Names["EmbeddedFiles"].Add(xRefTable.XRefTable, a.ID, *ir, m, []string{"F", "UF"})
	if err != nil {
		return err
	}

	rootDict, err := xRefTable.Catalog()
	if err != nil {
		return err
	}
	rootDict.Insert("AF", types.Array{*ir})
	return nil
}

// NewEmbeddedStreamDict creates and returns an embeddedStreamDict containing the bytes represented by r.
func (xRefTable *xRefTable) newEmbeddedStreamDict(r io.Reader, modDate time.Time) (*types.IndirectRef, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	bb := buf.Bytes()

	sd, err := xRefTable.NewStreamDictForBuf(bb)
	if err != nil {
		return nil, err
	}

	sd.InsertName("Type", "EmbeddedFile")
	sd.InsertName("Subtype", "text/xml")
	d := types.NewDict()
	d.InsertInt("Size", len(bb))
	d.Insert("ModDate", types.StringLiteral(types.DateString(modDate)))
	sd.Insert("Params", d)
	if err = sd.Encode(); err != nil {
		return nil, err
	}

	return xRefTable.IndRefForNewObject(*sd)
}

// NewEmbeddedStreamDict creates and returns an embeddedStreamDict containing the bytes represented by r.
func (xRefTable *xRefTable) newMetadataStreamDict(r io.Reader) (*types.IndirectRef, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, err
	}

	bb := buf.Bytes()

	sd := &types.StreamDict{
		Dict:    types.NewDict(),
		Content: bb,
	}
	sd.InsertName("Type", "Metadata")
	sd.InsertName("Subtype", "XML")
	if err := sd.Encode(); err != nil {
		return nil, err
	}

	return xRefTable.IndRefForNewObject(*sd)
}

func (xRefTable *xRefTable) newFileSpecDict(f, uf, desc, afRel string, indRefStreamDict types.IndirectRef) (types.Dict, error) {
	d := types.NewDict()
	d.InsertName("Type", "Filespec")

	s, err := types.EscapedUTF16String(f)
	if err != nil {
		return nil, err
	}
	d.InsertString("F", *s)

	if s, err = types.EscapedUTF16String(uf); err != nil {
		return nil, err
	}
	d.InsertString("UF", *s)

	d.InsertName("AFRelationship", afRel)

	efDict := types.NewDict()
	efDict.Insert("F", indRefStreamDict)
	efDict.Insert("UF", indRefStreamDict)
	d.Insert("EF", efDict)

	if desc != "" {
		if s, err = types.EscapedUTF16String(desc); err != nil {
			return nil, err
		}
		d.InsertString("Desc", *s)
	}

	// CI, optional, collection item dict, since V1.7
	// a corresponding collection schema dict in a collection.
	ciDict := types.NewDict()
	//add contextual meta info here.
	d.Insert("CI", ciDict)

	return d, nil
}
