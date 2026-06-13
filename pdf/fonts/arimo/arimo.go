package arimo

import (
	_ "embed"

	"github.com/signintech/gopdf"
)

//go:embed Arimo-Invoice-Regular.ttf
var regular []byte

//go:embed Arimo-Invoice-Italic.ttf
var italic []byte

func Setup(pdf *gopdf.GoPdf) error {
	err := pdf.AddTTFFontDataWithOption("arimo", regular, gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Regular,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return err
	}

	err = pdf.AddTTFFontDataWithOption("arimo", italic, gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Italic,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return err
	}
	err = pdf.SetFont("arimo", "", 12)
	if err != nil {
		return err
	}
	return nil
}
