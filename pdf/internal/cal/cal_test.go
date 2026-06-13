package cal

import (
	"testing"

	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/l10n"
)

var testInputs = []string{
	"DE 2026-01-02 2026-01-05 => 2.1. – 5.1.2026",
	"DE 2026-01-02 2026-03-05 => 2.1. – 5.3.2026",
	"DE 2025-10-02 2026-03-05 => 2.10.2025 – 5.3.2026",
	"DE 2026-03-01 2026-09-30 => März – Sept. 2026",
	"DE 2025-04-01 2026-09-30 => April 2025 – Sept. 2026",

	"GB 2026-01-02 2026-01-05 => Jan 2 – Jan 5, 2026",
	"GB 2026-01-02 2026-03-05 => Jan 2 – Mar 5, 2026",
	"GB 2025-10-02 2026-03-05 => Oct 2, 2025 – Mar 5, 2026",
	"GB 2026-03-01 2026-09-30 => Mar – Sep 2026",
	"GB 2025-04-01 2026-09-30 => Apr 2025 – Sep 2026",
}

func TestFormat(t *testing.T) {
	var period cal.Period
	for _, test := range testInputs {
		country := l10n.Code(test[:2])
		_ = period.Start.UnmarshalText([]byte(test[3 : 3+10]))
		_ = period.End.UnmarshalText([]byte(test[14 : 14+10]))
		expect := test[28:]
		got := FormatPeriod(&period, country)
		if got != expect {
			t.Errorf("expected: %q, got: %q", expect, got)
		}
	}
}
