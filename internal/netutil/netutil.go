package netutil

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/big"
	"net/netip"
)

// ParseIPNumber converts a raw base-10 IP number into an IP address. For example: 42540986356047386130018232263459733504 == 2001:1890:17dd:4500::
//
// For numbers less than the max uint32 an IPv4 address is returned, else an IPv6 address is returned.
func ParseIPNumber(rawIPNumber string) (addr netip.Addr, err error) {
	n, ok := big.NewInt(0).SetString(rawIPNumber, 10)
	if !ok {
		return addr, fmt.Errorf("invalid ip number")
	}

	if n.BitLen() <= 32 {
		var bs [4]byte
		n.FillBytes(bs[:])
		addr = netip.AddrFrom4(bs)
		return addr, nil
	}

	var bs [16]byte
	n.FillBytes(bs[:])
	addr = netip.AddrFrom16(bs)

	if addr.Is4In6() {
		return netip.AddrFrom4(addr.As4()), nil
	}

	return addr, nil
}

// AddrRangeToPrefixes converts an ip address range into a list of CIDR prefixes.
func AddrRangeToPrefixes(start, end netip.Addr) []netip.Prefix {
	// IPV4 -> IPV6 range is not supported
	if start.BitLen() != end.BitLen() {
		return nil
	}

	switch start.Compare(end) {
	case 0:
		// addresses are equal, so return a /32 (or /128) cidr
		prefix, _ := start.Prefix(start.BitLen())
		return []netip.Prefix{prefix}
	case 1:
		// end is before start, so no cidrs
		return nil
	}

	var prefixes []netip.Prefix
	for start.Compare(end) <= 0 {
		prefix, _ := start.Prefix(start.BitLen())
	loop:
		for i := start.BitLen() - 1; i >= 0; i-- {
			p, _ := start.Prefix(i)
			a1, a2 := PrefixToAddrRange(p)
			if a1 != start {
				break
			}
			switch a2.Compare(end) {
			case -1:
				prefix = p
			case 0:
				prefix = p
				break loop
			case 1:
				break loop
			}
		}
		prefixes = append(prefixes, prefix)
		_, start = PrefixToAddrRange(prefix)
		if start == end {
			break
		}
		start = start.Next()
	}
	return prefixes
}

// PrefixToAddrRange returns a CIDR prefix's inclusive ip address range
func PrefixToAddrRange(prefix netip.Prefix) (start, end netip.Addr) {
	start = prefix.Masked().Addr()
	end = transformAddr(start,
		func(ip *uint32) {
			shift := 32 - prefix.Bits()
			*ip |= (1 << shift) - 1
		},
		func(hi, lo *uint64) {
			shift := 128 - prefix.Bits()
			if shift >= 64 {
				shift -= 64
				*hi |= (1 << shift) - 1
				*lo = math.MaxUint64
			} else {
				*lo |= (1 << shift) - 1
			}
		},
	)
	return start, end
}

func transformAddr(addr netip.Addr, ipv4callback func(ip *uint32), ipv6callback func(hi, lo *uint64)) netip.Addr {
	if addr.Is4() {
		bs := addr.As4()
		ip := binary.BigEndian.Uint32(bs[:])
		ipv4callback(&ip)
		binary.BigEndian.PutUint32(bs[:], ip)
		return netip.AddrFrom4(bs)
	}

	bs := addr.As16()
	hi := binary.BigEndian.Uint64(bs[0:8])
	lo := binary.BigEndian.Uint64(bs[8:16])
	ipv6callback(&hi, &lo)
	binary.BigEndian.PutUint64(bs[0:8], hi)
	binary.BigEndian.PutUint64(bs[8:16], lo)
	return netip.AddrFrom16(bs)
}
