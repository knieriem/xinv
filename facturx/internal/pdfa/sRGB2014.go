package pdfa

import (
	_ "embed"
)

// sRGB2014.icc v2 from https://registry.color.org/rgb-registry/srgbprofiles
// Copyright International Color Consortium, 2015
// Licensing conditions: see https://registry.color.org/profile-library/#license
//
//	"This profile is made available by the International Color Consortium,
//	 and may be copied, distributed, embedded, made, used,
//	 and sold without restriction. Altered versions of this profile
//	 shall have the original identification and copyright information
//	 removed and shall not be misrepresented as the original profile."
//
//go:embed sRGB2014.icc
var iccData []byte

var OutputIntent_sRGB = OutputIntent{
	OutputCondition:           "RGB",
	OutputConditionIdentifier: "Custom",
	Profile: &ICCProfile{
		Data:               iccData,
		NumColorComponents: 3,
		Info:               "sRGB2014",
	},
}
