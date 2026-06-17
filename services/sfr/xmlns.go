package sfr

type Namespaces struct {
	NS  string `xml:"xmlns,attr,omitempty"`
	NS5 string `xml:"xmlns:ns5,attr,omitempty"`
	NS2 string `xml:"xmlns:ns2,attr,omitempty"`
	NS3 string `xml:"xmlns:ns3,attr,omitempty"`
	AF  string `xml:"xmlns:АФ,attr,omitempty"`
	UT  string `xml:"xmlns:УТ,attr,omitempty"`
	VZL string `xml:"xmlns:ВЗЛ,attr,omitempty"`
}
