package netstat

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"regexp"
	"strconv"

	"github.com/evilsocket/opensnitch/daemon/core"
	"github.com/evilsocket/opensnitch/daemon/log"
)

var (
	parser = regexp.MustCompile(`(?i)` +
		`\d+:\s+` + // sl
		`([a-f0-9]{8}):([a-f0-9]{4})\s+` + // local_address
		`([a-f0-9]{8}):([a-f0-9]{4})\s+` + // rem_address
		`[a-f0-9]{2}\s+` + // st
		`[a-f0-9]{8}:[a-f0-9]{8}\s+` + // tx_queue rx_queue
		`[a-f0-9]{2}:[a-f0-9]{8}\s+` + // tr tm->when
		`[a-f0-9]{8}\s+` + // retrnsmt
		`(\d+)\s+` + // uid
		`\d+\s+` + // timeout
		`(\d+)\s+` + // inode
		`.+`) // stuff we don't care about
)

func decToInt(n string) int {
	d, err := strconv.ParseInt(n, 10, 64)
	if err != nil {
		log.Fatal("Error while parsing %s to int: %s", n, err)
	}
	return int(d)
}

func hexToInt(h string) int {
	d, err := strconv.ParseInt(h, 16, 64)
	if err != nil {
		log.Fatal("Error while parsing %s to int: %s", h, err)
	}
	return int(d)
}

func hexToIP(h string) net.IP {
	n := hexToInt(h)
	ip := make(net.IP, 4)
	binary.LittleEndian.PutUint32(ip, uint32(n))
	return ip
}

func Parse(proto string) ([]Entry, error) {
	filename := fmt.Sprintf("/proc/net/%s", proto)
	fd, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	entries := make([]Entry, 0)
	scanner := bufio.NewScanner(fd)
	for lineno := 0; scanner.Scan(); lineno++ {
		// skip column names
		if lineno == 0 {
			continue
		}

		line := core.Trim(scanner.Text())
		m := parser.FindStringSubmatch(line)
		if m == nil {
			log.Warning("Could not parse netstat line from %s: %s", filename, line)
			continue
		}

		entries = append(entries, NewEntry(
			proto,
			hexToIP(m[1]),
			hexToInt(m[2]),
			hexToIP(m[3]),
			hexToInt(m[4]),
			decToInt(m[5]),
			decToInt(m[6]),
		))
	}

	return entries, nil
}
