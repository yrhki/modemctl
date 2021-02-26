package modem

import (
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

	token, err := c.getToken()
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

	token, err := c.getToken()
	if err != nil { return err }

	v := url.Values{
		"x.SourceIPStart":   {filter.SourceIPRange.StartString()},
		"x.SourceIPEnd":     {filter.SourceIPRange.EndString()},
		"x.DestIPStart":     {filter.DestIPRange.StartString()},
		"x.DestIPEnd":       {filter.DestIPRange.EndString()},
		"x.Protocol":        {filter.Protocol.string()},
		"x.SourcePortStart": {fmt.Sprint(filter.SourcePortRange.Start())},
		"x.SourcePortEnd":   {fmt.Sprint(filter.SourcePortRange.End())},
		"x.DestPortStart":   {fmt.Sprint(filter.DestPortRange.Start())},
		"x.DestPortEnd":     {fmt.Sprint(filter.DestPortRange.End())},
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


