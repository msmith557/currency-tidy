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
		{"€10,00", "€1,000.00"}, // no locale support yet: comma is a thousands separator
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

func TestParseErrors(t *testing.T) {
	for _, in := range []string{"", "   ", "abc", "$", "12.345"} {
		if _, err := Parse(in); err == nil {
			t.Errorf("Parse(%q): expected an error, got nil", in)
		}
	}
}
