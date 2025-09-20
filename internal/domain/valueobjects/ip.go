package valueobjects

import (
    "net"
)

type IP struct{ String string }

func NewIP(s string) (IP, bool) {
    ip := net.ParseIP(s)
    if ip == nil {
        return IP{}, false
    }
    return IP{String: ip.String()}, true
}
