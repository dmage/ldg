package parser

import (
	"fmt"
	"strings"
)

func isSpace(c byte) bool {
	return c == ' ' || c == '\t'
}

func indexSpace(s string) int {
	for i := 0; i < len(s); i++ {
		if isSpace(s[i]) {
			return i
		}
	}
	return -1
}

func indexNotSpace(s string) int {
	for i := 0; i < len(s); i++ {
		if !isSpace(s[i]) {
			return i
		}
	}
	return -1
}

func isEmptyLine(s string) bool {
	return indexNotSpace(s) == -1
}

func trimSpaceLeft(s string) string {
	idx := indexNotSpace(s)
	if idx == -1 {
		return ""
	}
	return s[idx:]
}

func trimSpaceRight(s string) string {
	i := len(s) - 1
	for ; i >= 0; i-- {
		if !isSpace(s[i]) {
			return s[:i+1]
		}
	}
	return ""
}

func parseHeader(line string) *Transaction {
	t := &Transaction{}
	space := indexSpace(line)
	if space == -1 {
		t.Date = line
		line = ""
	} else {
		t.Date = line[:space]
		line = trimSpaceLeft(line[space:])
	}
	if len(line) > 0 {
		if line[0] == '*' || line[0] == '!' {
			t.Status = rune(line[0])
			line = trimSpaceLeft(line[1:])
		}
	}
	t.Description = line
	return t
}

func parsePosting(line string) *Posting {
	account := ""
	space := -1
	i := 0
	for ; i < len(line); i++ {
		if line[i] == '\t' {
			account = trimSpaceRight(line[:i])
			line = line[i+1:]
			break
		}
		if line[i] == ' ' {
			if space == i-1 {
				account = line[:i-1]
				line = line[i+1:]
				break
			}
			space = i
		}
	}
	if account == "" {
		account = trimSpaceRight(line)
		line = ""
	}
	amount := trimSpaceLeft(line)
	directTag := ""
	i = strings.IndexByte(amount, ';')
	if i != -1 {
		directTag = amount[i:]
		amount = trimSpaceRight(amount[:i])
	} else {
		amount = trimSpaceRight(amount)
	}
	amountExtra := ""
	i = strings.IndexAny(amount, "@=")
	if i != -1 {
		amountExtra = amount[i:]
		amount = trimSpaceRight(amount[:i])
	}
	return &Posting{
		Account:     account,
		Amount:      amount,
		AmountExtra: amountExtra,
		DirectTag:   directTag,
	}
}

type Parser struct {
	input  string
	lineno int
	cur    string
	eof    bool

	transactions []*Transaction

	tr *Transaction
	p  *Posting
}

type stateFn func(p *Parser) (stateFn, error)

func (p *Parser) next() {
	idx := strings.IndexByte(p.input, '\n')
	var line string
	if idx == -1 {
		line = p.input
		p.input = ""
		p.eof = true
	} else {
		line = p.input[:idx+1]
		p.input = p.input[idx+1:]

	}
	if line != "" {
		p.lineno++
		p.cur = strings.TrimSuffix(line, "\n")
	}
}

func (p *Parser) Parse(content string) ([]*Transaction, error) {
	p.input = content
	p.lineno = 0
	p.eof = false
	p.transactions = nil
	state := parseFile
	for {
		p.next()
		var err error
		state, err = state(p)
		if err != nil {
			return p.transactions, err
		}
		if state == nil {
			break
		}
	}
	return p.transactions, nil
}

func parseFile(p *Parser) (stateFn, error) {
	if p.eof {
		return nil, nil
	}
	if isEmptyLine(p.cur) {
		return parseFile, nil
	}
	if isSpace(p.cur[0]) {
		return nil, fmt.Errorf("line %d: expected a transaction header, but got a posting", p.lineno)
	}
	p.tr = parseHeader(p.cur)
	p.p = nil
	p.transactions = append(p.transactions, p.tr)
	return parseTransaction, nil
}

func parseTransaction(p *Parser) (stateFn, error) {
	if p.eof {
		return nil, nil
	}
	if p.cur != "" && !isSpace(p.cur[0]) {
		return parseFile(p)
	}
	line := trimSpaceLeft(p.cur)
	if line == "" {
		return parseFile, nil
	}
	if line[0] == ';' {
		if p.p == nil {
			p.tr.Tags = append(p.tr.Tags, line)
		} else {
			p.p.ExtraTags = append(p.p.ExtraTags, line)
		}
		return parseTransaction, nil
	}
	p.p = parsePosting(line)
	p.tr.Postings = append(p.tr.Postings, p.p)
	return parseTransaction, nil
}

func New() *Parser {
	return &Parser{}
}

func Parse(content string) ([]*Transaction, error) {
	p := New()
	return p.Parse(content)
}
