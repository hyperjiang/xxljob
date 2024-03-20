package xxljob

import (
	"fmt"
	"net"
	"time"
)

// LocalIP gets local IPv4 address.
func LocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// ReadableSize prints the size in human readable format.
func ReadableSize(size int64) string {
	s := float64(size)
	if size >= 1024*1024*1024 {
		return fmt.Sprintf("%.02fgb", s/1024/1024/1024)
	} else if size >= 1024*1024 {
		return fmt.Sprintf("%.02fmb", s/1024/1024)
	} else if size >= 1024 {
		return fmt.Sprintf("%.02fkb", s/1024)
	}

	return fmt.Sprintf("%db", size)
}

// TruncateDuration truncates a duration less than 1s to the specified precision,
// otherwise returns the duration unchanged.
func TruncateDuration(d time.Duration) time.Duration {
	if d >= time.Second {
		return d
	}

	if d >= time.Millisecond {
		return d.Truncate(time.Millisecond)
	}

	return d.Truncate(time.Microsecond)
}
