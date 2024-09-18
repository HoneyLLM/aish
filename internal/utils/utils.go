package utils

import (
	"bufio"
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
