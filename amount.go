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

// symbolCurrency maps a currency symbol to the one ISO code Parse assigns
// it. Several of these symbols are shared by multiple currencies in real
// use (¥ for both JPY and CNY, $ for USD, CAD, AUD, and others); each is
// pointed at whichever currency is the more common source of amounts
// without an explicit ISO code, since a symbol alone can't disambiguate
// further. Anything not listed here still parses fine as long as it's
// tagged with its three-letter ISO code instead of a symbol.
var symbolCurrency = map[string]string{
	"$": "USD",
	"£": "GBP",
	"€": "EUR",
	"¥": "JPY",
	"₹": "INR",
	"₩": "KRW",
	"₽": "RUB",
	"₺": "TRY",
	"₫": "VND",
	"₪": "ILS",
	"₴": "UAH",
	"₦": "NGN",
	"R$": "BRL",
}

var currencySymbol = map[string]string{
	"USD": "$",
	"GBP": "£",
	"EUR": "€",
	"JPY": "¥",
	"INR": "₹",
	"KRW": "₩",
	"RUB": "₽",
	"TRY": "₺",
	"VND": "₫",
	"ILS": "₪",
	"UAH": "₴",
	"NGN": "₦",
	"BRL": "R$",
}

// Parse turns a messy amount string into an Amount. It accepts a leading or
// trailing currency symbol or three-letter ISO code, thousands separators,
// surrounding whitespace, and a negative value written with a leading minus
// or wrapped in parentheses (accounting notation).
//
// Both the US convention (comma thousands, period decimal) and the European
// convention (period thousands, comma decimal) are recognized. When only one
// kind of separator appears, the one before the final 1 or 2 digits is taken
// as the decimal point, per European usage ("10,00", "1.234,5"); a lone
// separator followed by three digits is assumed to be a thousands grouping
// ("1,234", "$1.234") since that's the far more common case in practice, and
// no real currency amount has a three-digit fractional part anyway.
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

	s = normalizeSeparators(s)
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

// CurrencyStyle selects how Format marks the currency of the amounts it
// renders.
type CurrencyStyle int

const (
	// StyleSymbol renders a known currency symbol ("$", "€") and falls back
	// to the ISO code when no symbol is registered for it. This is the
	// default and matches Format's long-standing behavior.
	StyleSymbol CurrencyStyle = iota
	// StyleISOCode always renders the three-letter ISO code ("USD", "EUR"),
	// even for currencies that have a symbol.
	StyleISOCode
	// StyleNone omits the currency marker entirely, leaving just the signed,
	// grouped number. Useful when the currency is already known from
	// context, such as a single-currency CSV column.
	StyleNone
)

// Format renders an Amount as a canonical string: a currency symbol when
// one is known (otherwise the ISO code), comma-grouped thousands, exactly
// two decimal places, and a leading minus for negative values.
func (a Amount) Format() string {
	return a.FormatStyle(StyleSymbol)
}

// FormatStyle renders an Amount like Format, but with the currency marker
// controlled by style instead of always preferring a symbol.
func (a Amount) FormatStyle(style CurrencyStyle) string {
	minor := a.Minor
	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}

	whole := minor / 100
	frac := minor % 100
	number := groupThousands(strconv.FormatInt(whole, 10)) + "." + pad2(frac)

	switch style {
	case StyleISOCode:
		return sign + a.Currency + " " + number
	case StyleNone:
		return sign + number
	default:
		sym, ok := currencySymbol[a.Currency]
		if !ok {
			sym = a.Currency + " "
		}
		return sign + sym + number
	}
}

// normalizeSeparators rewrites whichever separator is acting as the decimal
// point in s to '.' and strips whichever is acting as a thousands grouping,
// so the rest of Parse only ever has to deal with a single '.' decimal.
func normalizeSeparators(s string) string {
	lastComma := strings.LastIndexByte(s, ',')
	lastDot := strings.LastIndexByte(s, '.')

	switch {
	case lastComma == -1:
		// No comma at all: dot, if any, is already the decimal point.
		return s
	case lastDot == -1:
		// Comma-only: the last comma is the decimal point if it's followed
		// by 1 or 2 digits, otherwise every comma is a thousands separator.
		if fracLen := len(s) - lastComma - 1; fracLen == 1 || fracLen == 2 {
			return strings.ReplaceAll(s[:lastComma], ",", "") + "." + s[lastComma+1:]
		}
		return strings.ReplaceAll(s, ",", "")
	case lastComma > lastDot:
		// Comma comes after the dot ("1.234,56"): comma is the decimal
		// point, every dot is a thousands separator.
		return strings.Replace(strings.ReplaceAll(s, ".", ""), ",", ".", 1)
	default:
		// Dot comes after the comma ("1,234.56"): the familiar US form.
		return strings.ReplaceAll(s, ",", "")
	}
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
