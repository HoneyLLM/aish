package utils

import (
	"bufio"
	"context"
	"io"
	"iter"
	"os"
	"strings"
)

func GetWriter(path string) (io.Writer, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

// ReadShellCommand continuously reads from the provided bufio.Reader,
// parses shell commands considering quotes and escaped characters,
// and yields each complete command along with any errors encountered.
func ReadShellCommand(r *bufio.Reader) iter.Seq2[string, error] {
	return func(yield func(string, error) bool) {
		var (
			builder       strings.Builder
			inSingleQuote bool
			inDoubleQuote bool
			escaped       bool
		)

		for {
			runeChar, _, err := r.ReadRune()
			if err != nil {
				if err == io.EOF {
					// If there's any remaining command in the builder, yield it
					if builder.Len() > 0 {
						if !yield(builder.String(), nil) {
							return
						}
					}
					return
				}
				// Yield the error and stop
				yield("", err)
				return
			}

			char := runeChar

			if escaped {
				// If the previous character was an escape, add this character and reset escape
				builder.WriteRune(char)
				escaped = false
				continue
			}

			if char == '\\' {
				// Next character is escaped
				escaped = true
				continue
			}

			if char == '\'' && !inDoubleQuote {
				// Toggle single quote state
				inSingleQuote = !inSingleQuote
				builder.WriteRune(char)
				continue
			}

			if char == '"' && !inSingleQuote {
				// Toggle double quote state
				inDoubleQuote = !inDoubleQuote
				builder.WriteRune(char)
				continue
			}

			if char == '\n' && !inSingleQuote && !inDoubleQuote {
				// End of command
				command := builder.String()
				// Reset builder for next command
				builder.Reset()
				if !yield(command, nil) {
					return
				}
				continue
			}

			// Regular character
			builder.WriteRune(char)
		}
	}
}

// ContextReader wraps an io.Reader and adds context support.
type ContextReader struct {
	ctx    context.Context
	reader io.Reader
}

// NewContextReader creates a new ContextReader.
// Note that it may read an extra chunk after cancelled.
func NewContextReader(ctx context.Context, reader io.Reader) *ContextReader {
	return &ContextReader{
		ctx:    ctx,
		reader: reader,
	}
}

// Read reads data into p, respecting the context's cancellation or deadline.
func (cr *ContextReader) Read(p []byte) (n int, err error) {
	type readResult struct {
		n   int
		err error
	}

	resultCh := make(chan readResult, 1)

	// Start a goroutine to perform the read operation.
	go func() {
		n, err := cr.reader.Read(p)
		resultCh <- readResult{n: n, err: err}
	}()

	select {
	case <-cr.ctx.Done():
		return 0, cr.ctx.Err()
	case res := <-resultCh:
		return res.n, res.err
	}
}
