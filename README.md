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

## Current limitations

- No locale awareness: a comma is always read as a thousands separator
  and a period as the decimal point, so European-style `10,00` reads as
  1000, not 10.
- Currency detection is symbol- or ISO-code-based only; amounts with no
  marker at all default to USD.
- A single line longer than 1 MiB is rejected rather than streamed, to
  keep an adversarial no-newline input from growing memory unbounded.

## License

MIT, see [LICENSE](LICENSE).
