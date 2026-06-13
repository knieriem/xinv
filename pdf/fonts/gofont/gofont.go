package gofont

import (
	"github.com/signintech/gopdf"

	"golang.org/x/image/font/gofont/goitalic"
	"golang.org/x/image/font/gofont/goregular"
)

func Setup(pdf *gopdf.GoPdf) error {
	err := pdf.AddTTFFontDataWithOption("lato", goregular.TTF, gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Regular,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return err
	}

	err = pdf.AddTTFFontDataWithOption("lato", goitalic.TTF, gopdf.TtfOption{
		UseKerning:                true,
		Style:                     gopdf.Italic,
		OnGlyphNotFoundSubstitute: gopdf.DefaultOnGlyphNotFoundSubstitute,
	})
	if err != nil {
		return err
	}
	err = pdf.SetFont("lato", "", 28)
	if err != nil {
		return err
	}
	return nil
}
