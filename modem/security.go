package modem

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"regexp"
	"strconv"
	"strings"

)

type IPFilterProtocol int

var (
	regIpInfo = regexp.MustCompile(`IpInfo\("(Internet.*)","(.*)","(.*)","(.*)","(.*)","(.*)","(.*)","(.*)","(.*)","(.*)"\)`)
	regBlackList = regexp.MustCompile(`var BlackListInfo = new Array\((.*),null\);`)
	regWhiteList = regexp.MustCompile(`var WhiteListInfo = new Array\((.*),null\);`)
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
	DestIPRange, SourceIPRange [2]net.IP
	DestPortRange, SourcePortRange [2]int
}


func (ipf *IPFilter) ID() string {
	if ipf.block {
		return fmt.Sprintf("")
	} else {
		return fmt.Sprintf("")
	}
}

func (ipf *IPFilter) sourceIP(pos int) string {
	if ipf.SourceIPRange[pos] != nil {
		return ipf.SourceIPRange[pos].String()
	}
	return "0"
}

func (ipf *IPFilter) destIP(pos int) string {
	if ipf.DestIPRange[pos] != nil {
		return ipf.DestIPRange[pos].String()
	}
	return "0"
}





func parseCIRD(cidr string) (net.IP, net.IP, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil { return nil, nil, err }

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

	return ipFirst, ipLast, nil
}

func IPFilterCIDR(sourceCIDR, destCIDR string) (*IPFilter, error) {
	filter := new(IPFilter)

	if sourceCIDR != "" {
		first, last, err := parseCIRD(sourceCIDR)
		if err != nil { return nil, err }
		filter.SourceIPRange = [2]net.IP{first, last}
	}

	if destCIDR != "" {
		first, last, err := parseCIRD(destCIDR)
		if err != nil { return nil, err }
		filter.DestIPRange = [2]net.IP{first, last}
	}

	return filter, nil
}

func parseIpInfo(reg *regexp.Regexp, text string) ([]*IPFilter, error) {
	filterList := []*IPFilter{}

	if match := reg.FindStringSubmatch(text); len(match) > 0 {
		list := match[1]
		for _, info := range strings.Split(list, ",new") {
			fields := regIpInfo.FindStringSubmatch(info)

			srcPs, err := strconv.Atoi(fields[7])
			if err != nil { panic(err) }
			srcPe, err := strconv.Atoi(fields[8])
			if err != nil { panic(err) }
			dstPs, err := strconv.Atoi(fields[9])
			if err != nil { panic(err) }
			dstPe, err := strconv.Atoi(fields[10])
			if err != nil { panic(err) }

			f := IPFilter{
				// TODO: id: fields[1],
				SourceIPRange: [2]net.IP{net.ParseIP(fields[2]), net.ParseIP(fields[3])},
				SourcePortRange: [2]int{srcPs, srcPe},
				DestIPRange: [2]net.IP{net.ParseIP(fields[4]), net.ParseIP(fields[5])},
				DestPortRange: [2]int{dstPs, dstPe},
			}

			switch fields[6] {
			case "ALL":
				f.Protocol = IPFALL
			case "TCP":
				f.Protocol = IPFTCP
			case "UDP":
				f.Protocol = IPFUDP
			case "TCP/UDP":
				f.Protocol = IPFTCPUDP
			case "ICMP":
				f.Protocol = IPFICMP
			default:
				panic("Unexpected protocol: " + fields[6])
			}

			filterList = append(filterList, &f)
		}
	}
	return filterList, nil
}

func (c *Client) GetIPFilters() (black, white []*IPFilter, err error) {
	resp, err := c.c.Get(c.formatURL("/html/security/ipfilter.asp"))
	if err != nil { return nil, nil, err }
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil { return nil, nil, err }

	listBlack, err := parseIpInfo(regBlackList, string(b))
	if err != nil { return nil, nil, err }
	listWhite, err := parseIpInfo(regWhiteList, string(b))
	if err != nil { return nil, nil, err }

	return listBlack, listWhite, nil
}

func (c *Client) DeleteIPFilter(filters []*IPFilter) error {
	if len(filters) == 0 { return nil }

	token, err := c.getToken("/html/security/ipfilter.asp")
	if err != nil { return err }

	v := token.form()

	for _, filter := range filters {
		v[filter.ID()] = []string{""}
	}

	_, err = c.httpPostForm("/html/security/delFw.cgi?RequestFile=success&fwstat=2", v)
	if err != nil { return err }

	return nil
}

func (c *Client) DeleteAllIPFilter() error {
	b, w, err := c.GetIPFilters()
	if err != nil { return err }

	err = c.DeleteIPFilter(b)
	if err != nil { return err }

	err = c.DeleteIPFilter(w)
	if err != nil { return err }

	return nil
}




// 
// 'list' true 'Blacklist' false 'Whitelist'
// IPFilter ID takes precedence over 'list'
func (c *Client) AddIPFilter(block bool, filter *IPFilter) error {
	// TODO: Rename function to make more sense

	token, err := c.getToken("/html/security/ipfilter.asp")
	if err != nil { return err }

	v := url.Values{
		"x.SourceIPStart":   {filter.sourceIP(0)},
		"x.SourceIPEnd":     {filter.sourceIP(1)},
		"x.DestIPStart":     {filter.destIP(0)},
		"x.DestIPEnd":       {filter.destIP(1)},
		"x.Protocol":        {filter.Protocol.string()},
		"x.SourcePortStart": {fmt.Sprint(filter.SourcePortRange[0])},
		"x.SourcePortEnd":   {fmt.Sprint(filter.SourcePortRange[1])},
		"x.DestPortStart":   {fmt.Sprint(filter.DestPortRange[0])},
		"x.DestPortEnd":     {fmt.Sprint(filter.DestPortRange[0])},
		"csrf_param":        {token.csrfParam},
		"csrf_token":        {token.csrfToken},
	}

	var listType string

	if block {
		listType = "IpBlackFilter"
	} else {
		listType = "IpWhiteFilter"
	}

	fmt.Printf("%#+v\n", v)

	if filter.ID() == "" {
		// Add new
		_, err = c.httpPostForm("/html/security/addcfgFw.cgi?x=InternetGatewayDevice.X_FireWall." + listType, v)
		if err != nil { return err }
	} else {
		// Update using id
		_, err = c.httpPostForm("/html/security/addcfgFw.cgi?x=" + filter.ID(), v)
		if err != nil { return err }
	}

	return nil
}


