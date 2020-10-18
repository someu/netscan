package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	ipSplice     = "([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])"
	cidrMask     = "([1-9]|[1-2]\\d|3[0-2])"
	cidrReString = fmt.Sprintf("^(%s\\.){3}%s\\/%s$", ipSplice, ipSplice, cidrMask)
	CidrRe       = regexp.MustCompile(cidrReString)
	ipReString   = fmt.Sprintf("^(%s\\.){3}%s$", ipSplice, ipSplice)
	IPRe         = regexp.MustCompile(ipReString)
	portReString = "(\\d{1,4}|5\\d{4}|6[0-4]\\d{3}|65[0-4]\\d{2}|655[0-2]\\d|6553[0-5])"
	PortRe       = regexp.MustCompile(fmt.Sprintf("^%s(-%s)?$", portReString, portReString))
)

var PrimeRootTable = [][2]uint{
	{3, 2},
	{5, 2},
	{17, 3},
	{97, 5},
	{193, 5},
	{257, 3}, // 2^8 + 1
	{7681, 17},
	{12289, 11},
	{40961, 3},
	{65537, 3}, // 2^16 + 1
	{786433, 10},
	{5767169, 3},
	{7340033, 3},
	{16777259, 2}, // 2^24 + 43
	{23068673, 3},
	{104857601, 3},
	{167772161, 3},
	{268435459, 2}, // 2^28 + 3
	{469762049, 3},
	{1004535809, 3},
	{2013265921, 31},
	{2281701377, 3},
	{3221225473, 5},
	{4294967311, 3}, // 2^32 + 15
}

// number segment, range is [start, end), length is "end + 1 - start"
type Segment struct {
	start  uint
	end    uint
	length uint
}

type Segments []*Segment

func (segs Segments) get(i uint) (uint, error) {
	for _, seg := range segs {
		if seg.length >= i {
			return seg.start + i, nil
		} else {
			i -= seg.length
		}
	}
	return 0, errors.New("Index is out of range in segments\n")
}

type TargetRange struct {
	ipSegments     Segments
	portSegments   Segments
	ipCount        uint
	portCount      uint
	totalCount     uint
	prime          uint
	primeRoot      uint
	currentIndex   uint // range of index is [0, totalCount - 1]
	currentInverse uint // range of inverse is [1, totalCount]
}

func NewTargetRange(ipSegments Segments, portSegments Segments) *TargetRange {
	var ipCount uint
	for _, seg := range ipSegments {
		ipCount += seg.length
	}
	var portCount uint
	for _, seg := range portSegments {
		portCount += seg.length
	}
	var totalCount = ipCount * portCount

	var prime uint
	var primeRoot uint
	for _, pr := range PrimeRootTable {
		if pr[0] >= totalCount {
			prime = pr[0]
			primeRoot = pr[1]
			break
		}
	}

	targetRange := &TargetRange{
		ipSegments:   ipSegments,
		portSegments: portSegments,
		ipCount:      ipCount,
		portCount:    portCount,
		totalCount:   totalCount,
		prime:        prime,
		primeRoot:    primeRoot,
	}

	_ = targetRange.setCurrent(0)

	return targetRange
}

func (r *TargetRange) setCurrent(current uint) error {
	if current >= r.totalCount {
		return errors.New("CurrentIndex is out of range in target ranger\n")
	}
	var i uint = 1
	var inverse uint = 1
	for ; i <= current; i++ {
		inverse = (inverse * r.primeRoot) % r.prime
	}
	r.currentIndex = current
	r.currentInverse = inverse
	return nil
}

func (r TargetRange) hasNext() bool {
	return r.currentIndex < r.totalCount
}

func (r *TargetRange) nextTarget() (net.IP, uint16, error) {
	if r.currentIndex >= r.totalCount {
		return nil, 0, errors.New("Index is out of range in target range\n")
	}
	// ensure the range of inverse is match the range of index
	var inverse = r.currentInverse - 1

	// calc next inverse in cyclic group
	for {
		r.currentInverse = (r.currentInverse * r.primeRoot) % r.prime
		r.currentIndex++
		if r.currentInverse <= r.totalCount {
			break
		}
	}

	ipLong, err := r.ipSegments.get(inverse % r.ipCount)
	if err != nil {
		return nil, 0, err
	}

	portLong, err := r.portSegments.get(inverse / r.ipCount)
	if err != nil {
		return nil, 0, err
	}
	return longToIP(uint32(ipLong)), uint16(portLong), nil
}

func parseIpSegment(ipStr string) (*Segment, error) {
	if IPRe.MatchString(ipStr) {
		ipLong := ipToLong(net.ParseIP(ipStr))
		return &Segment{
			start:  uint(ipLong),
			end:    uint(ipLong) + 1,
			length: 1,
		}, nil
	} else if CidrRe.MatchString(ipStr) {
		_, cidr, _ := net.ParseCIDR(ipStr)

		start := uint(ipToLong(cidr.IP))
		length := uint(^uint32(0) & ^binary.BigEndian.Uint32(cidr.Mask))

		return &Segment{
			start:  start,
			end:    start + length - 1,
			length: length,
		}, nil
	}
	return nil, errors.New("ipStr is not in ip or cidr format")
}

func parsePortSegment(portStr string) (*Segment, error) {
	if !PortRe.MatchString(portStr) {
		return nil, errors.New("portStr is not in port format")
	}
	i := strings.LastIndex(portStr, "-")
	var start int
	var end int
	if i >= 0 {
		start, _ = strconv.Atoi(portStr[0:i])
		end, _ = strconv.Atoi(portStr[i+1 : 0])
	} else {
		start, _ = strconv.Atoi(portStr)
		end = start
	}

	if start <= end {
		return &Segment{
			start:  uint(start),
			end:    uint(end),
			length: uint(end - start + 1),
		}, nil
	} else {
		return &Segment{
			start:  uint(end),
			end:    uint(start),
			length: uint(start - end + 1),
		}, nil
	}
}
