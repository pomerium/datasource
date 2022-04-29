package netutil

import (
	"errors"
	"net/netip"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseIPNumber(t *testing.T) {
	for _, testCase := range []struct {
		raw    string
		expect netip.Addr
		err    error
	}{
		{"0", netip.MustParseAddr("0.0.0.0"), nil},
		{"4294967295", netip.MustParseAddr("255.255.255.255"), nil},
		{"4294967296", netip.MustParseAddr("::1:0:0"), nil},
		{"42540986356047386130018232263459733504", netip.MustParseAddr("2001:1890:17dd:4500::"), nil},
		{"340282366920938463463374607431768211455", netip.MustParseAddr("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"), nil},
		{"2001:1890:17dd:4500::", netip.Addr{}, errors.New("invalid ip number")},
	} {
		actual, err := ParseIPNumber(testCase.raw)
		assert.Equal(t, testCase.err, err)
		assert.Equal(t, testCase.expect.String(), actual.String())
	}
}

func TestAddrRangeToPrefixes(t *testing.T) {
	for _, testCase := range []struct {
		start, end, expect string
	}{
		{"192.168.0.1", "192.168.0.1", "192.168.0.1/32"},
		{"192.168.0.0", "192.168.0.255", "192.168.0.0/24"},
		{"192.168.0.96", "192.168.0.255", "192.168.0.96/27,192.168.0.128/25"},
	} {
		start := netip.MustParseAddr(testCase.start)
		end := netip.MustParseAddr(testCase.end)
		actual := AddrRangeToPrefixes(start, end)
		var actualStrs []string
		for _, p := range actual {
			actualStrs = append(actualStrs, p.String())
		}
		assert.Equal(t, testCase.expect, strings.Join(actualStrs, ","),
			"start=%s end=%s", start, end)
	}
}

func TestPrefixToAddrRange(t *testing.T) {
	for _, testCase := range []struct {
		prefix, expectedStart, expectedEnd string
	}{
		{"10.0.0.0/8", "10.0.0.0", "10.255.255.255"},
		{"fc00::/7", "fc00::", "fdff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"},
		{"::ffff:0:0/96", "::ffff:0.0.0.0", "::ffff:255.255.255.255"},
	} {
		prefix := netip.MustParsePrefix(testCase.prefix)
		start, end := PrefixToAddrRange(prefix)
		assert.Equal(t, testCase.expectedStart, start.String(),
			"prefix=%s", prefix)
		assert.Equal(t, testCase.expectedEnd, end.String(),
			"prefix=%s", prefix)
	}
}
