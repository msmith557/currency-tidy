// Command normalize reads currency amounts from stdin, one per line, and
// writes the normalized form of each to stdout. Lines it can't parse are
// passed through with a "# " prefix instead of stopping the whole run.
package main

import (
	"fmt"
	"os"

	tidy "github.com/msmith557/currency-tidy"
)

func main() {
	if err := tidy.Stream(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "normalize:", err)
		os.Exit(1)
	}
}
