package inspector

import (
	"strconv"
	"strings"
)

func parseStatus(input string) map[string]string {
	lines := strings.Split(input, "\n")
	out := make(map[string]string)
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			out[key] = val
		}
	}
	return out
}

func atoi(s string) int {
	i, _ := strconv.Atoi(strings.Fields(s)[0]) // take only first part
	return i
}
