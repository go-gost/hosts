package hosts

import (
	"bufio"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

// Host is a static mapping from hostname to IP.
type Host struct {
	IP       net.IP
	Hostname string
	Aliases  []string
}

// NewHost creates a Host.
func NewHost(ip net.IP, hostname string, aliases ...string) Host {
	return Host{
		IP:       ip,
		Hostname: hostname,
		Aliases:  aliases,
	}
}

// Hosts is an interface that performs static table lookup for host name.
type Hosts interface {
	Lookup(host string) net.IP
}

// hosts is a static table lookup for hostnames.
// For each host a single line should be present with the following information:
// IP_address canonical_hostname [aliases...]
// Fields of the entry are separated by any number of blanks and/or tab characters.
// Text from a "#" character until the end of the line is a comment, and is ignored.
type staticHosts struct {
	hosts   []Host
	period  time.Duration
	stopped chan struct{}
	mux     sync.RWMutex
}

// NewHosts creates a Hosts with optional list of hosts.
func NewHosts(hosts ...Host) Hosts {
	return &staticHosts{
		hosts:   hosts,
		stopped: make(chan struct{}),
	}
}

// Lookup searches the IP address corresponds to the given host from the host table.
func (h *staticHosts) Lookup(host string) (ip net.IP) {
	if h == nil || host == "" {
		return
	}

	h.mux.RLock()
	defer h.mux.RUnlock()

	for _, h := range h.hosts {
		if h.Hostname == host {
			ip = h.IP
			break
		}
		for _, alias := range h.Aliases {
			if alias == host {
				ip = h.IP
				break
			}
		}
	}
	return
}

// Reload parses config from r, then live reloads the hosts.
func (h *staticHosts) Reload(r io.Reader) error {
	var period time.Duration
	var hosts []Host

	if r == nil || h.Stopped() {
		return nil
	}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		ss := splitLine(line)
		if len(ss) < 2 {
			continue // invalid lines are ignored
		}

		switch ss[0] {
		case "reload": // reload option
			period, _ = time.ParseDuration(ss[1])
		default:
			ip := net.ParseIP(ss[0])
			if ip == nil {
				break // invalid IP addresses are ignored
			}
			host := Host{
				IP:       ip,
				Hostname: ss[1],
			}
			if len(ss) > 2 {
				host.Aliases = ss[2:]
			}
			hosts = append(hosts, host)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	h.mux.Lock()
	h.period = period
	h.hosts = hosts
	h.mux.Unlock()

	return nil
}

// Period returns the reload period
func (h *staticHosts) Period() time.Duration {
	if h.Stopped() {
		return -1
	}

	h.mux.RLock()
	defer h.mux.RUnlock()

	return h.period
}

// Stop stops reloading.
func (h *staticHosts) Stop() {
	select {
	case <-h.stopped:
	default:
		close(h.stopped)
	}
}

// Stopped checks whether the reloader is stopped.
func (h *staticHosts) Stopped() bool {
	select {
	case <-h.stopped:
		return true
	default:
		return false
	}
}

// splitLine splits a line text by white space, mainly used by config parser.
func splitLine(line string) []string {
	if line == "" {
		return nil
	}
	if n := strings.IndexByte(line, '#'); n >= 0 {
		line = line[:n]
	}
	line = strings.Replace(line, "\t", " ", -1)
	line = strings.TrimSpace(line)

	var ss []string
	for _, s := range strings.Split(line, " ") {
		if s = strings.TrimSpace(s); s != "" {
			ss = append(ss, s)
		}
	}
	return ss
}
