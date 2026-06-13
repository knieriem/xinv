package ti

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/knieriem/text/tidata"
)

func ParseFile(filename string, data any) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return Parse(f, data)
}

func Parse(r io.Reader, conf interface{}) (err error) {
	el, err := readTiData(r)
	if err != nil {
		return
	}

	err = el.Decode(conf, &ticonf)
	if err != nil {
		return
	}
	return
}

func readTiData(r io.Reader) (el *tidata.Elem, err error) {
	tr := tidata.NewReader(bufio.NewScanner(r))
	tr.CommentPrefix = "#"
	tr.CommentPrefixEscaped = `\#`
	el, err = tr.ReadAll()
	return
}

var ticonf = tidata.Config{
	Sep:    ":",
	MapSym: ":",
	KeyToFieldName: func(key string) (name string) {
		switch key {
		case "sepa":
			return "SEPA"
		case "vat":
			return "VAT"
		case "iban":
			return "IBAN"
		}
		s := strings.Title(key)
		s = replaceSpecial(s, "-", "")
		name = s
		if strings.HasSuffix(name, "Id") {
			name = name[:len(name)-1] + "D"
		} else if strings.HasSuffix(name, "Url") {
			name = name[:len(name)-2] + "RL"
		}
		return
	},
	MultiStringSep: "\n",
}

func replaceSpecial(s, old, new string) string {
	f := strings.Split(s, old)
	if len(f) == 0 {
		return s
	}
	s = f[0]
	for _, seg := range f[1:] {
		s += new + strings.Title(seg)
	}
	return s
}
