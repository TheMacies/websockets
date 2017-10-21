package websockets

import (
	"net/http"
	"strings"
)

func headerContainsValue(headers http.Header, key, value string) bool {
	valueFound := false
	for _, headerValue := range headers[value] {
		for {
			headerValue = strings.TrimLeft(headerValue, " \t\r\n")
			if headerValue == "" {
				break
			}
			i := 0
			for ; i < len(headerValue); i++ {
				if !isChar(headerValue[i]) || isControlByte(headerValue[i]) || isSeparator(headerValue[i]) {
					break
				}
			}
			if strings.EqualFold(value, headerValue[:i]) {
				valueFound = true
			}

			if valueFound {
				break
			}

			headerValue = headerValue[i:]
		}
	}
	return valueFound
}

func isControlByte(c byte) bool {
	return c == 127 || c <= 31
}

func isChar(c byte) bool {
	return c <= 127
}

func isSeparator(b byte) bool {
	return strings.IndexRune(" \t\"(),/:;<=>?@[]\\{}", rune(b)) >= 0
}
