# currency-tidy

Amounts that come from spreadsheets, invoices, CSV exports, and pasted
emails are never in one consistent format. Some have a `$`, some have
`USD` after the number, some use parentheses for negative values, some
have thousands commas and some don't. Before you can add, compare, or
store these amounts you need to normalize them, and it's easy to get the
edge cases (accounting-style negatives, currency on the wrong side,
values missing cents) subtly wrong.

`currency-tidy` parses that mess into an `Amount` (an integer count of
minor units plus a currency code, never a float) and formats it back out
consistently.

```go
package main

import (
	"fmt"

	tidy "github.com/msmith557/currency-tidy"
)

func main() {
	for _, raw := range []string{"$1,234.56", "(99.5)", "USD 40", "-12"} {
		amt, err := tidy.Parse(raw)
		if err != nil {
			fmt.Println(raw, "-> error:", err)
			continue
		}
		fmt.Println(raw, "->", amt.Format())
	}
}
```

```
$1,234.56 -> $1,234.56
(99.5) -> -$99.50
USD 40 -> $40.00
-12 -> -$12.00
```

## Streaming

Amount lists show up as large files: a year of transaction exports, a
batch of invoices. `Stream` reads one line at a time and writes the
normalized result immediately, so memory use stays flat no matter how
big the input is — it never buffers more than the current line.

```go
package main

import (
	"os"

	tidy "github.com/msmith557/currency-tidy"
)

func main() {
	tidy.Stream(os.Stdin, os.Stdout)
}
```

A small CLI wrapper is included:

```
go run ./cmd/normalize < messy_amounts.txt > clean_amounts.txt
```

Lines that don't parse are written through unchanged with a `# ` prefix
instead of stopping the run, so one bad row in a 10 million line file
doesn't cost you the other 9,999,999.

By default the currency is rendered as a symbol (`$40.00`). Pass `-style`
to change that:

```
go run ./cmd/normalize -style=iso < messy_amounts.txt   # USD 40.00
go run ./cmd/normalize -style=none < messy_amounts.txt  # 40.00
```

The same options are available from the library through `tidy.WithStyle`,
passed to `Stream`, or `Amount.FormatStyle` for one-off values.

## Current limitations

- Both US-style (`1,234.56`) and European-style (`1.234,56`) separators are
  recognized, but only one at a time: Parse looks at the string in front of
  it and infers which convention it's in, it isn't told the locale up
  front.
- Currency detection is symbol- or ISO-code-based only; amounts with no
  marker at all default to USD.
- A single line longer than 1 MiB is rejected rather than streamed, to
  keep an adversarial no-newline input from growing memory unbounded.

## License

MIT, see [LICENSE](LICENSE).
