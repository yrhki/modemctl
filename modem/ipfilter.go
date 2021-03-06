package modem

import (
	"encoding/binary"
	"fmt"
	"net"
)


const (
	IPFALL IPFilterProtocol = iota
	IPFTCP
	IPFUDP
	IPFTCPUDP
	IPFICMP
)

func (ipfp IPFilterProtocol) string() string {
	switch ipfp {
	case IPFALL:
		return "ALL"
	case IPFTCP:
		return "TCP"
	case IPFUDP:
		return "UDP"
	case IPFTCPUDP:
		return "TCP/UDP"
	case IPFICMP:
		return "ICMP"
	default:
		panic("IPFilterProtocol unexpected value")
	}
}

type IPFilter struct {
	block bool
	id int
	Protocol IPFilterProtocol
	DestIPRange, SourceIPRange IPRange
	DestPortRange, SourcePortRange PortRange
}

func (ipf *IPFilter) ID() string {
	if ipf.id != 0 {
		if ipf.block {
			return fmt.Sprintf("InternetGatewayDevice.X_FireWall.IpBlackFilter.%d", ipf.id)
		} else {
			return fmt.Sprintf("InternetGatewayDevice.X_FireWall.IpWhiteFilter.%d", ipf.id)
		}
	}
	return ""
}

func parseCIRD(cidr string) (IPRange, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil { return IPRange{}, err }

	// https://stackoverflow.com/questions/60540465/go-how-to-list-all-ips-in-a-network/60542265#60542265
	mask := binary.BigEndian.Uint32(ipNet.Mask)
	start := binary.BigEndian.Uint32(ipNet.IP)

	// make last bit 254 instead of 255
	last := (start & mask) | (mask ^ 0xfffffffe)
	ipLast := make(net.IP, 4)
	binary.BigEndian.PutUint32(ipLast, last)

	// make starting address end in 1
	ipFirst := make(net.IP, 4)
	if start % 8 == 0 { start++ }
	binary.BigEndian.PutUint32(ipFirst, start)

	return IPRange{ipFirst, ipLast}, nil
}

func IPFilterCIDR(sourceCIDR, destCIDR string) (*IPFilter, error) {
	filter := new(IPFilter)

	if sourceCIDR != "" {
		ipRange, err := parseCIRD(sourceCIDR)
		if err != nil { return nil, err }
		filter.SourceIPRange = ipRange
	}

	if destCIDR != "" {
		ipRange, err := parseCIRD(destCIDR)
		if err != nil { return nil, err }
		filter.DestIPRange = ipRange
	}

	return filter, nil
}
