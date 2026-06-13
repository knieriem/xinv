package cal

import (
	"fmt"
	"strings"

	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/l10n"
)

type Date = cal.Date

type Period = cal.Period

type formatting struct {
	monthNames       []string
	ymdFmt           string
	mdFmt            string
	yearFirst        bool
	preferMonthNames bool
	omitSameYear     bool
	commaBeforeYear  bool
}

var fmtmap = map[l10n.Code]*formatting{
	l10n.DE: {
		ymdFmt:       "2.1.2006",
		mdFmt:        "2.1.",
		omitSameYear: true,
		monthNames: []string{
			"n/a", "Jan.", "Feb.", "März", "April",
			"Mai", "Juni", "Juli", "Aug.",
			"Sept.", "Okt.", "Nov.", "Dez.",
		},
	},
	l10n.GB: {
		preferMonthNames: true,
		commaBeforeYear:  true,
		monthNames: []string{
			"n/a", "Jan", "Feb", "Mar", "Apr",
			"May", "Jun", "Jul", "Aug",
			"Sep", "Oct", "Nov", "Dec",
		},
	},
}

const enDash = "\u2013"

func FormatDate(d *Date, lang l10n.Code) string {
	fm, ok := fmtmap[lang]
	if !ok {
		fm = fmtmap[l10n.GB]
	}

	cy := ""
	if fm.commaBeforeYear {
		cy = ","
	}

	if fm.preferMonthNames {
		return fmt.Sprintf("%s %d%s %d", fm.monthNames[d.Month], d.Day, cy, d.Year)
	}
	return fmt.Sprintf("%v", d.Time().Format(fm.ymdFmt))
}

func FormatPeriod(p *Period, lang l10n.Code) string {
	endFullMonth := p.End.Time().AddDate(0, 0, 1).Day() == 1
	fullMonths := p.Start.Day == 1 && endFullMonth
	sameYear := p.Start.Year == p.End.Year

	fm, ok := fmtmap[lang]
	if !ok {
		fm = fmtmap[l10n.GB]
	}

	cy := ""
	if fm.commaBeforeYear {
		cy = ","
	}

	sb := new(strings.Builder)
	if fullMonths {
		fmt.Fprintf(sb, "%s", fm.monthNames[p.Start.Month])
		if !sameYear {
			fmt.Fprintf(sb, " %d", p.Start.Year)
		}
		fmt.Fprintf(sb, " %s %s %d", enDash, fm.monthNames[p.End.Month], p.End.Year)
		return sb.String()
	}
	if fm.preferMonthNames {
		fmt.Fprintf(sb, "%s %d", fm.monthNames[p.Start.Month], p.Start.Day)
		if !sameYear {
			fmt.Fprintf(sb, "%s %d", cy, p.Start.Year)
		}
		fmt.Fprintf(sb, " %s %s %d%s %d", enDash, fm.monthNames[p.End.Month], p.End.Day, cy, p.End.Year)
		return sb.String()
	}

	date1Fmt := fm.ymdFmt
	if sameYear {
		date1Fmt = fm.mdFmt
	}
	fmt.Fprintf(sb, "%v", p.Start.Time().Format(date1Fmt))
	fmt.Fprintf(sb, " %v %v", enDash, p.End.Time().Format(fm.ymdFmt))
	return sb.String()
}
