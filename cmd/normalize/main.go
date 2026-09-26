// Command normalize reads currency amounts from stdin, one per line, and
// writes the normalized form of each to stdout. Lines it can't parse are
// passed through with a "# " prefix instead of stopping the whole run.
package main

import (
	"flag"
	"fmt"
	"os"

	tidy "github.com/msmith557/currency-tidy"
)

func main() {
	style := flag.String("style", "symbol", "currency marker style: symbol, iso, or none")
	flag.Parse()

	var opt tidy.Option
	switch *style {
	case "symbol":
		opt = tidy.WithStyle(tidy.StyleSymbol)
	case "iso":
		opt = tidy.WithStyle(tidy.StyleISOCode)
	case "none":
		opt = tidy.WithStyle(tidy.StyleNone)
	default:
		fmt.Fprintf(os.Stderr, "normalize: unknown -style %q (want symbol, iso, or none)\n", *style)
		os.Exit(2)
	}

	if err := tidy.Stream(os.Stdin, os.Stdout, opt); err != nil {
		fmt.Fprintln(os.Stderr, "normalize:", err)
		os.Exit(1)
	}
}
