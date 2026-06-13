module github.com/knieriem/xinv

go 1.25.0

require (
	github.com/bojanz/currency v1.4.4
	github.com/invopop/gobl v0.401.0
	github.com/invopop/gobl.cii v0.35.0
	github.com/pdfcpu/pdfcpu v0.13.0
	github.com/signintech/gopdf v0.33.0
	golang.org/x/image v0.42.0
	golang.org/x/text v0.38.0
)

require (
	cloud.google.com/go v0.123.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20230301143203-a9d515a09cc2 // indirect
	github.com/bahlo/generic-list-go v0.2.0 // indirect
	github.com/buger/jsonparser v1.2.0 // indirect
	github.com/clipperhouse/uax29/v2 v2.7.0 // indirect
	github.com/cockroachdb/apd/v3 v3.2.3 // indirect
	github.com/expr-lang/expr v1.17.8 // indirect
	github.com/go-jose/go-jose/v4 v4.1.4 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/hhrutter/lzw v1.0.0 // indirect
	github.com/hhrutter/pkcs7 v0.2.2 // indirect
	github.com/hhrutter/tiff v1.0.3 // indirect
	github.com/invopop/jsonschema v0.14.0 // indirect
	github.com/invopop/validation v0.8.0 // indirect
	github.com/invopop/xmlctx v0.13.0 // indirect
	github.com/invopop/yaml v0.3.1 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/magefile/mage v1.17.2 // indirect
	github.com/mattn/go-runewidth v0.0.24 // indirect
	github.com/pb33f/ordered-map/v2 v2.3.1 // indirect
	github.com/phpdave11/gofpdi v1.0.14-0.20211212211723-1f10f9844311 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.yaml.in/yaml/v4 v4.0.0-rc.5 // indirect
	golang.org/x/crypto v0.52.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/signintech/gopdf v0.33.0 => github.com/knieriem/gopdf v0.0.0-20250924094535-bf364b20e7f7

replace github.com/pdfcpu/pdfcpu v0.13.0 => github.com/knieriem/pdfcpu v0.0.0-20260609213303-f9cf18c4bc96

replace github.com/invopop/gobl.cii v0.35.0 => github.com/knieriem/gobl.cii v0.0.0-20260607204932-f9568feb3b1d
