package tidy

import "testing"

func TestParseFormat(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"$1,234.56", "$1,234.56"},
		{"1234.56", "$1,234.56"},
		{"-$12", "-$12.00"},
		{"$-12", "-$12.00"},
		{"(12.34)", "-$12.34"},
		{"1234.5", "$1,234.50"},
		{"USD 99.99", "$99.99"},
		{"5", "$5.00"},
		{"€10,00", "€10.00"},              // European decimal comma
		{"€1.234,56", "€1,234.56"},        // European thousands dot + decimal comma
		{"1.234.567,89", "$1,234,567.89"}, // European thousands, no currency marker so USD default
		{"1,234", "$1,234.00"},            // lone comma before 3 digits: still a thousands separator
		{"¥1,234.56", "¥1,234.56"},
		{"₹99.99", "₹99.99"},
		{"R$40", "R$40.00"},
		{"KRW 500", "₩500.00"}, // ISO code in, matching symbol out
		{"-₩500", "-₩500.00"},
		{"CHF 40", "CHF 40.00"}, // ISO code with no symbol mapping falls back to the code itself
	}

	for _, c := range cases {
		amt, err := Parse(c.in)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", c.in, err)
			continue
		}
		if got := amt.Format(); got != c.want {
			t.Errorf("Parse(%q).Format() = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestFormatStyle(t *testing.T) {
	cases := []struct {
		in    string
		style CurrencyStyle
		want  string
	}{
		{"$1,234.56", StyleSymbol, "$1,234.56"},
		{"$1,234.56", StyleISOCode, "USD 1,234.56"},
		{"$1,234.56", StyleNone, "1,234.56"},
		{"-$12", StyleISOCode, "-USD 12.00"},
		{"-$12", StyleNone, "-12.00"},
		{"CHF 40", StyleISOCode, "CHF 40.00"},
		{"CHF 40", StyleNone, "40.00"},
	}

	for _, c := range cases {
		amt, err := Parse(c.in)
		if err != nil {
			t.Fatalf("Parse(%q) returned error: %v", c.in, err)
		}
		if got := amt.FormatStyle(c.style); got != c.want {
			t.Errorf("Parse(%q).FormatStyle(%v) = %q, want %q", c.in, c.style, got, c.want)
		}
	}
}

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "abc", "$", "12.345"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q): expected an error, got nil", in)
		}
	}
}
