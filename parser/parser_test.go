package parser

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseHeaader(t *testing.T) {
	var testCases = []struct {
		input string
		want  *Transaction
	}{
		{
			input: "2022/05/12",
			want: &Transaction{
				Date: "2022/05/12",
			},
		},
		{
			input: "2022/05/12 * Hello world",
			want: &Transaction{
				Date:        "2022/05/12",
				Status:      '*',
				Description: "Hello world",
			},
		},
		{
			input: "2022/05/12*",
			want: &Transaction{
				Date: "2022/05/12*",
			},
		},
		{
			input: "2022/05/12 *",
			want: &Transaction{
				Date:   "2022/05/12",
				Status: '*',
			},
		},
		{
			input: "2022/05/12 !Hello",
			want: &Transaction{
				Date:        "2022/05/12",
				Status:      '!',
				Description: "Hello",
			},
		},
		{
			input: "2022/05/12 Test",
			want: &Transaction{
				Date:        "2022/05/12",
				Description: "Test",
			},
		},
		{
			input: "2022-05-12\t    *   \t Test",
			want: &Transaction{
				Date:        "2022-05-12",
				Status:      '*',
				Description: "Test",
			},
		},
	}
	for _, tc := range testCases {
		got := parseHeader(tc.input)
		if got.Date != tc.want.Date {
			t.Errorf("parseHeader(%q).Date = %q, want %q", tc.input, got.Date, tc.want.Date)
		}
		if got.Status != tc.want.Status {
			t.Errorf("parseHeader(%q).Status = %q, want %q", tc.input, got.Status, tc.want.Status)
		}
		if got.Description != tc.want.Description {
			t.Errorf("parseHeader(%q).Description = %q, want %q", tc.input, got.Description, tc.want.Description)
		}
	}
}

func TestParsePosting(t *testing.T) {
	var testCases = []struct {
		input string
		want  *Posting
	}{
		{
			input: "Assets:Cash",
			want: &Posting{
				Account: "Assets:Cash",
			},
		},
		{
			input: "Assets:Cash  50.00 EUR",
			want: &Posting{
				Account: "Assets:Cash",
				Amount:  "50.00 EUR",
			},
		},
		{
			input: "Assets:Cash   50.00 EUR",
			want: &Posting{
				Account: "Assets:Cash",
				Amount:  "50.00 EUR",
			},
		},
		{
			input: "Assets:Cash    50.00 EUR",
			want: &Posting{
				Account: "Assets:Cash",
				Amount:  "50.00 EUR",
			},
		},
		{
			input: "Assets:Cash     50.00 EUR",
			want: &Posting{
				Account: "Assets:Cash",
				Amount:  "50.00 EUR",
			},
		},
		{
			input: "Assets:Cash\t50.00 EUR",
			want: &Posting{
				Account: "Assets:Cash",
				Amount:  "50.00 EUR",
			},
		},
		{
			input: "Assets:Cash  50.00 EUR;comment",
			want: &Posting{
				Account: "Assets:Cash",
				Amount:  "50.00 EUR",
			},
		},
	}
	for _, tc := range testCases {
		got := parsePosting(tc.input)
		if got.Account != tc.want.Account {
			t.Errorf("parsePosting(%q).Account = %q, want %q", tc.input, got.Account, tc.want.Account)
		}
		if got.Amount != tc.want.Amount {
			t.Errorf("parsePosting(%q).Amount = %q, want %q", tc.input, got.Amount, tc.want.Amount)
		}
	}
}

func TestParse(t *testing.T) {
	var testCases = []struct {
		input string
		want  []*Transaction
	}{
		{
			input: `2022/05/12 * Supermarket
    Expenses:Food:Groceries                10.00 EUR
    Assets:Cash
`,
			want: []*Transaction{
				{
					Date:        "2022/05/12",
					Status:      '*',
					Description: "Supermarket",
					Postings: []*Posting{
						{
							Account: "Expenses:Food:Groceries",
							Amount:  "10.00 EUR",
						},
						{
							Account: "Assets:Cash",
						},
					},
				},
			},
		},
		{
			input: `2022/05/12 * Supermarket
    ; ID: 12345
    Expenses:Food:Groceries  10.00 EUR  ; :FOO:
    Assets:Cash  ; :BAR:
`,
			want: []*Transaction{
				{
					Date:        "2022/05/12",
					Status:      '*',
					Description: "Supermarket",
					Tags: []string{
						"; ID: 12345",
					},
					Postings: []*Posting{
						{
							Account:   "Expenses:Food:Groceries",
							Amount:    "10.00 EUR",
							DirectTag: "; :FOO:",
						},
						{
							Account:   "Assets:Cash",
							DirectTag: "; :BAR:",
						},
					},
				},
			},
		},
	}
	for _, tc := range testCases {
		got, err := Parse(tc.input)
		if err != nil {
			t.Errorf("Parse(%q) failed: %v", tc.input, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			diff := cmp.Diff(got, tc.want)
			t.Errorf("Parse(%q):\n%s", tc.input, diff)
		}
	}
}

func BenchmarkParser(b *testing.B) {
	const content = `2022/05/12 * Supermarket
    Expenses:Food:Groceries                10.00 EUR
    Assets:Cash
`
	p := New()
	for n := 0; n < b.N; n++ {
		_, err := p.Parse(content)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func serializeTransactions(transactions []*Transaction) string {
	var s string
	for i, t := range transactions {
		if i != 0 {
			s += "\n"
		}
		s += t.String()
	}
	return s
}

func TestFmt(t *testing.T) {
	files, err := filepath.Glob("./testdata/fmt/*.in.ldg")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		t.Run(strings.TrimPrefix(file, "testdata/fmt/"), func(t *testing.T) {
			t.Parallel()
			input, err := ioutil.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			transactions, err := Parse(string(input))
			if err != nil {
				t.Fatal(err)
			}
			got := serializeTransactions(transactions)
			want, err := ioutil.ReadFile(strings.TrimSuffix(file, ".in.ldg") + ".out.ldg")
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Errorf("Parse(%q) = %q, want %q", input, got, want)
			}
		})
	}
}

func FuzzParse(f *testing.F) {
	f.Add("2022/05/12 * Test\n    A  1 USD\n    B\n")
	f.Fuzz(func(t *testing.T, input string) {
		transactions, err := Parse(input)
		if err != nil {
			return
		}
		buf, _ := json.Marshal(transactions)
		canonical := serializeTransactions(transactions)
		transactions, err = Parse(canonical)
		if err != nil {
			t.Errorf("Parse(%q): %v", canonical, err)
		}
		got := serializeTransactions(transactions)
		if got != canonical {
			t.Errorf("input:  %q", input)
			t.Errorf("transactions: %s", buf)
			t.Errorf("first:  %q", canonical)
			t.Errorf("second: %q", got)
		}
	})
}
