package utils

import (
	"bytes"
	"io"
	"os"
	"strings"
)

func GetWriter(path string) (io.Writer, error) {
	return os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
}

func GetLastLine(s string) string {
	lastIndex := strings.LastIndex(s, "\n")
	if lastIndex == -1 {
		return s
	}
	return s[lastIndex+1:]
}

func RemoveLastLine(s string) string {
	lastIndex := strings.LastIndex(s, "\n")
	if lastIndex == -1 {
		return ""
	}
	return s[:lastIndex+1]
}

func StringToLineChannel(strChan <-chan string) <-chan string {
	lineChan := make(chan string)

	go func() {
		defer close(lineChan)

		var buffer []byte
		for {
			str, ok := <-strChan
			if !ok {
				if len(buffer) > 0 {
					lineChan <- string(buffer)
				}
				return
			}

			buffer = append(buffer, str...)

			for {
				idx := bytes.IndexByte(buffer, '\n')
				if idx == -1 {
					break
				}

				line := string(buffer[:idx])
				buffer = buffer[idx+1:]

				lineChan <- line
			}
		}
	}()

	return lineChan
}

func HandleChannel[T any](ch <-chan T, onEach func(T, bool), onFinish func([]T)) {
	var buffer []T
	var items []T

	for item := range ch {
		if len(buffer) > 0 {
			onEach(buffer[0], false)
			items = append(items, buffer[0])
			buffer = buffer[:0]
		}

		buffer = append(buffer, item)
	}

	if len(buffer) > 0 {
		onEach(buffer[0], true)
		items = append(items, buffer[0])
	}

	onFinish(items)
}
