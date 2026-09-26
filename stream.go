package tidy

import (
	"bufio"
	"io"
	"strings"
)

// maxLineSize bounds how large a single line is allowed to be. It exists so
// a pathological input (no newlines at all) can't force unbounded growth of
// the scanner's internal buffer.
const maxLineSize = 1 << 20 // 1 MiB

// Option configures a call to Stream.
type Option func(*streamConfig)

type streamConfig struct {
	style CurrencyStyle
}

// WithStyle sets the currency style Stream uses when formatting each parsed
// line. The default, when no Option is given, is StyleSymbol.
func WithStyle(style CurrencyStyle) Option {
	return func(c *streamConfig) { c.style = style }
}

// Stream reads amounts one line at a time from r, normalizes each with
// Parse and Format, and writes the result to w. Only the current line is
// ever held in memory, so Stream can process input far larger than
// available RAM without the caller having to chunk it themselves.
//
// A line that fails to parse is passed through unchanged, prefixed with
// "# ", so a caller scanning the output can spot it; Stream keeps going
// rather than aborting on the first bad line.
func Stream(r io.Reader, w io.Writer, opts ...Option) error {
	cfg := streamConfig{style: StyleSymbol}
	for _, opt := range opts {
		opt(&cfg)
	}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	bw := bufio.NewWriter(w)
	defer bw.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		amt, err := Parse(line)
		if err != nil {
			if _, err := bw.WriteString("# " + line + "\n"); err != nil {
				return err
			}
			continue
		}
		if _, err := bw.WriteString(amt.FormatStyle(cfg.style) + "\n"); err != nil {
			return err
		}
	}
	return scanner.Err()
}
