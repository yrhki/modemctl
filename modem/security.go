package modem

import (
	"fmt"
	"net/url"
)

type IPFilterProtocol int

func (c *Client) GetIPFilters() (block, allow []*IPFilter, err error) {
	config, err := c.GetConfig()
	if err != nil { return nil, nil, err }
	return config.Firewall.IPBlockFilters, config.Firewall.IPAllowFilters, nil
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


