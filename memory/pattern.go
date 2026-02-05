package memory

import (
	"fmt"
	"strconv"
	"strings"
)

type Pattern struct {
	Bytes []byte
	Mask  []bool
}

func ParsePattern(pattern string) (Pattern, error) {
	tokens := strings.Fields(pattern)
	if len(tokens) == 0 {
		return Pattern{}, fmt.Errorf("pattern is empty")
	}
	bytes := make([]byte, 0, len(tokens))
	mask := make([]bool, 0, len(tokens))
	for _, token := range tokens {
		if token == "?" || token == "??" || strings.Contains(token, "?") {
			bytes = append(bytes, 0)
			mask = append(mask, false)
			continue
		}
		if len(token) != 2 {
			return Pattern{}, fmt.Errorf("invalid token: %s", token)
		}
		val, err := strconv.ParseUint(token, 16, 8)
		if err != nil {
			return Pattern{}, fmt.Errorf("invalid token: %s", token)
		}
		bytes = append(bytes, byte(val))
		mask = append(mask, true)
	}
	return Pattern{Bytes: bytes, Mask: mask}, nil
}
