// Package tidy normalizes messy currency amount strings ("$1,234.56",
// "-12.5", "(99.00)", "USD 40") into a canonical Amount and back into a
// consistent string form.
package tidy

import (
	"errors"
	"strconv"
	"strings"
)

// Amount holds a currency value as an integer count of minor units (cents
// for USD) rather than a float, so formatting never has to round.
type Amount struct {
	Minor    int64
	Currency string
}

var (
	ErrEmpty  = errors.New("tidy: empty amount")
	ErrFormat = errors.New("tidy: unrecognized amount format")
)

var symbolCurrency = map[string]string{
	"$": "USD",
	"£": "GBP",
	"€": "EUR",
}

var currencySymbol = map[string]string{
	"USD": "$",
	"GBP": "£",
	"EUR": "€",
}

// Parse turns a messy amount string into an Amount. It accepts a leading or
// trailing currency symbol or three-letter ISO code, comma thousands
// separators, surrounding whitespace, and a negative value written with a
// leading minus or wrapped in parentheses (accounting notation).
//
// It does not attempt locale detection: a comma is always treated as a
// thousands separator and a period as the decimal point.
func Parse(raw string) (Amount, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Amount{}, ErrEmpty
	}

	negative := false
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		negative = true
		s = strings.TrimSpace(s[1 : len(s)-1])
	}
	if strings.HasPrefix(s, "-") {
		negative = true
		s = strings.TrimSpace(s[1:])
	} else if strings.HasPrefix(s, "+") {
		s = strings.TrimSpace(s[1:])
	}

	currency := ""
	for sym, code := range symbolCurrency {
		if strings.HasPrefix(s, sym) {
			currency = code
			s = strings.TrimSpace(s[len(sym):])
			break
		}
		if strings.HasSuffix(s, sym) {
			currency = code
			s = strings.TrimSpace(s[:len(s)-len(sym)])
			break
		}
	}
	if currency == "" {
		if fields := strings.Fields(s); len(fields) == 2 {
			if code, ok := isISOCode(fields[0]); ok {
				currency, s = code, fields[1]
			} else if code, ok := isISOCode(fields[1]); ok {
				currency, s = code, fields[0]
			}
		}
	}

	// The sign can sit on either side of a symbol ("$-12" as well as
	// "-$12"), so check again now that the symbol is gone.
	if strings.HasPrefix(s, "-") {
		negative = true
		s = s[1:]
	} else if strings.HasPrefix(s, "+") {
		s = s[1:]
	}

	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return Amount{}, ErrFormat
	}

	whole, frac, hasFrac := strings.Cut(s, ".")
	if whole == "" {
		whole = "0"
	}
	if !isDigits(whole) {
		return Amount{}, ErrFormat
	}

	if hasFrac {
		if len(frac) == 0 || len(frac) > 2 || !isDigits(frac) {
			return Amount{}, ErrFormat
		}
		if len(frac) == 1 {
			frac += "0"
		}
	} else {
		frac = "00"
	}

	wholeVal, err := strconv.ParseInt(whole, 10, 63)
	if err != nil {
		return Amount{}, ErrFormat
	}
	fracVal, err := strconv.ParseInt(frac, 10, 63)
	if err != nil {
		return Amount{}, ErrFormat
	}

	minor := wholeVal*100 + fracVal
	if negative {
		minor = -minor
	}
	if currency == "" {
		currency = "USD"
	}

	return Amount{Minor: minor, Currency: currency}, nil
}

// Format renders an Amount as a canonical string: a currency symbol when
// one is known (otherwise the ISO code), comma-grouped thousands, exactly
// two decimal places, and a leading minus for negative values.
func (a Amount) Format() string {
	minor := a.Minor
	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}

	whole := minor / 100
	frac := minor % 100

	sym, ok := currencySymbol[a.Currency]
	if !ok {
		sym = a.Currency + " "
	}

	return sign + sym + groupThousands(strconv.FormatInt(whole, 10)) + "." + pad2(frac)
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isISOCode(s string) (string, bool) {
	if len(s) != 3 {
		return "", false
	}
	for _, r := range s {
		if (r < 'A' || r > 'Z') && (r < 'a' || r > 'z') {
			return "", false
		}
	}
	return strings.ToUpper(s), true
}

func groupThousands(digits string) string {
	n := len(digits)
	if n <= 3 {
		return digits
	}
	var b strings.Builder
	lead := n % 3
	if lead > 0 {
		b.WriteString(digits[:lead])
	}
	for i := lead; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(digits[i : i+3])
	}
	return b.String()
}

func pad2(n int64) string {
	s := strconv.FormatInt(n, 10)
	if len(s) < 2 {
		return "0" + s
	}
	return s
}
