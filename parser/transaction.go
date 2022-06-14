package parser

import (
	"fmt"
	"io"
	"strings"
)

type Posting struct {
	Account     string
	Amount      string
	AmountExtra string
	DirectTag   string
	ExtraTags   []string
}

type Transaction struct {
	Date        string
	Status      rune
	Description string
	Tags        []string
	Postings    []*Posting
}

func (tr Transaction) write(b io.StringWriter) {
	status := ""
	if tr.Status != 0 {
		status = fmt.Sprintf("%c ", tr.Status)
	}
	b.WriteString(fmt.Sprintf("%s %s%s\n", tr.Date, status, tr.Description))
	for _, tag := range tr.Tags {
		b.WriteString(fmt.Sprintf("    %s\n", tag))
	}
	for _, p := range tr.Postings {
		if p.Amount != "" || p.AmountExtra != "" {
			amount := p.Amount
			amountExtra := ""
			if p.AmountExtra != "" {
				amountExtra = " " + p.AmountExtra
			}
			if p.DirectTag != "" {
				amountExtra += "  " + p.DirectTag
			}

			var accountWidth int
			if len(p.Account)+2 < 48-len(amount) {
				accountWidth = 48 - len(amount)
			} else {
				accountWidth = len(p.Account) + 2
			}
			// assertion: accountWidth >= len(p.Account) + 2

			b.WriteString(fmt.Sprintf("    %[1]*[2]s%s%s\n", -accountWidth, p.Account, amount, amountExtra))
		} else {
			amountExtra := ""
			if p.DirectTag != "" {
				amountExtra += "  " + p.DirectTag
			}
			b.WriteString(fmt.Sprintf("    %s%s\n", p.Account, amountExtra))
		}
		for _, tag := range p.ExtraTags {
			b.WriteString(fmt.Sprintf("    %s\n", tag))
		}
	}
}

func (tr Transaction) String() string {
	var b strings.Builder
	tr.write(&b)
	return b.String()
}
