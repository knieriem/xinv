package pdfa

import (
	_ "embed"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

type OutputIntent struct {
	OutputCondition           string
	OutputConditionIdentifier string
	Registry                  string
	Profile                   *ICCProfile
}

type ICCProfile struct {
	Data               []byte
	NumColorComponents int
	Info               string
}

func SetOutputIntent(ctx *model.Context, oi *OutputIntent) error {
	xRefTable := ctx.XRefTable

	prof := oi.Profile
	sd, err := xRefTable.NewStreamDictForBuf(prof.Data)
	if err != nil {
		return err
	}
	sd.InsertInt("N", prof.NumColorComponents)
	if err = sd.Encode(); err != nil {
		return err
	}

	sdRef, err := xRefTable.IndRefForNewObject(*sd)
	if err != nil {
		return err
	}

	d := types.NewDict()
	d.InsertName("Type", "OutputIntent")
	d.InsertName("S", "GTS_PDFA1")
	d.InsertString("OutputCondition", oi.OutputCondition)
	d.InsertString("OutputConditionIdentifier", oi.OutputConditionIdentifier)
	d.Insert("DestOutputProfile", *sdRef)
	d.InsertString("Info", prof.Info)
	oiRef, err := xRefTable.IndRefForNewObject(d)
	if err != nil {
		return err
	}

	rootDict, err := xRefTable.Catalog()
	if err != nil {
		return err
	}
	rootDict.Update("OutputIntents", types.Array{*oiRef})
	return nil
}
