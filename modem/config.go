package modem

import (
	"encoding/xml"
	"fmt"
	"net"
	"os"
	"strconv"
)

type configIPFilter struct {
	ID string `xml:"InstanceID,attr"`
}

type PortRange [2]int
func (r PortRange) Start() int { return r[0] }
func (r PortRange) End() int { return r[1] }

type IPRange [2]net.IP
func (r IPRange) Start() net.IP { return r[0] }
func (r IPRange) End() net.IP { return r[1] }

type Config struct {
	Firewall ConfigFirewall `xml:"InternetGatewayDevice>X_FireWall"`
	WANDevice ConfigWANDevice `xml:"InternetGatewayDevice>WANDevice"`
}

type ConfigFirewall struct {
	CurrentLevel string `xml:"CurrentLevel,attr"`
	BlockDOS int `xml:"BlockDoS,attr"`
	IPBlockFilters []*IPFilter `xml:"IpBlackFilter>IpBlackFilterInstance"`
	IPAllowFilters []*IPFilter `xml:"IpBlackFilter>IpWhiteFilterInstance"`
}

type ConfigWANDevice struct {
	NumberOfInstances int `xml:"NumberOfInstances,attr"`
	PortMaps []*PortMap `xml:"WANDeviceInstance>WANConnectionDevice>WANConnectionDeviceInstance>WANIPConnection>WANIPConnectionInstance>PortMapping>PortMappingInstance"`
}

func (pm *PortMap) UnmarshalXML(dec *xml.Decoder, s xml.StartElement) error {
	var err error

	pm.id, err = strconv.ParseUint(s.Attr[0].Value, 0, 0)
	if err != nil { return err }

	pm.Enabled = s.Attr[1].Value == "1"
	pm.RemoteHost, pm.LocalHost = net.ParseIP(s.Attr[2].Value), net.ParseIP(s.Attr[7].Value)
	pm.ExternalPortRange, err = parsePortRange(s.Attr[3], s.Attr[4])
	if err != nil { return err }
	pm.LocalPort, err = strconv.Atoi(s.Attr[5].Value)
	if err != nil { return err }
	pm.Description = s.Attr[8].Value

	switch s.Attr[6].Value {
	case "TCP":
		pm.Protocol = PortMapProtocolTCP
	case "UDP":
		pm.Protocol = PortMapProtocolUDP
	case "TCP/UDP":
		pm.Protocol = PortMapProtocolTCPUDP
	}

	_, _ = dec.Token()
	return nil
}

func (ipf *IPFilter) UnmarshalXML(dec *xml.Decoder, s xml.StartElement) error {
	var err error

	ipf.block = s.Name.Local == "IpBlackFilterInstance"

	ipf.id, err = strconv.Atoi(s.Attr[0].Value)
	if err != nil { return err }

	ipf.SourceIPRange = [2]net.IP{net.ParseIP(s.Attr[2].Value), net.ParseIP(s.Attr[3].Value)}
	ipf.DestIPRange   = [2]net.IP{net.ParseIP(s.Attr[4].Value), net.ParseIP(s.Attr[5].Value)}

	ipf.Protocol = parseIPFilterProtocol(s.Attr[6].Value)

	ipf.SourcePortRange, err = parsePortRange(s.Attr[7], s.Attr[8])
	if err != nil { return err }
	ipf.DestPortRange, err = parsePortRange(s.Attr[9], s.Attr[10])
	if err != nil { return err }
	_, _ = dec.Token()
	return nil
}

func parsePortRange(start, end xml.Attr) (PortRange, error) {
	var (
		values PortRange
		err error
	)

	values[0], err = strconv.Atoi(start.Value)
	if err != nil { return values, err }
	values[1], err = strconv.Atoi(end.Value)
	if err != nil { return values, err }

	return values, nil
}

func parseIPFilterProtocol(text string) IPFilterProtocol {
	switch text {
	case "ALL":
		return IPFALL
	case "TCP":
		return IPFTCP
	case "UDP":
		return IPFUDP
	case "TCP/UDP":
		return IPFTCPUDP
	case "ICMP":
		return IPFICMP
	default:
		panic("Unexpected protocol: " + text)
	}
}

func Parsetest() error {
	b, err := os.ReadFile("./config.xml")
	if err != nil { return err }

	data := new(Config)
	err = xml.Unmarshal(b, data)
	if err != nil { return err }

	// fmt.Printf("%#+v\n", data)

	fmt.Printf("%#+v\n", data)

	for _, pmap := range data.WANDevice.PortMaps {
		fmt.Println(pmap)
	}
	

	return nil
}
