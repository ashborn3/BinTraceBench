package inspector

import (
	"encoding/hex"
	"fmt"
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

func parseIPPort(hexstr string) string {
	parts := strings.Split(hexstr, ":")
	if len(parts) != 2 {
		return "?"
	}
	ipHex, portHex := parts[0], parts[1]
	port, _ := strconv.ParseInt(portHex, 16, 64)

	var ip string
	if len(ipHex) == 8 { // IPv4
		bs, _ := hex.DecodeString(ipHex)
		ip = fmt.Sprintf("%d.%d.%d.%d", bs[3], bs[2], bs[1], bs[0])
	} else if len(ipHex) == 32 { // IPv6
		ip = ipHex // simplified, can improve later
	} else {
		ip = "?"
	}
	return fmt.Sprintf("%s:%d", ip, port)
}

func tcpState(hex string) string {
	states := map[string]string{
		"01": "ESTABLISHED",
		"02": "SYN_SENT",
		"03": "SYN_RECV",
		"04": "FIN_WAIT1",
		"05": "FIN_WAIT2",
		"06": "TIME_WAIT",
		"07": "CLOSE",
		"08": "CLOSE_WAIT",
		"09": "LAST_ACK",
		"0A": "LISTEN",
		"0B": "CLOSING",
	}
	return states[strings.ToUpper(hex)]
}
