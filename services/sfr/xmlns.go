package sfr

type Namespaces struct {
	NS  string `xml:"xmlns,attr,omitempty"`
	NS2 string `xml:"xmlns:ns2,attr,omitempty"`
	AF  string `xml:"xmlns:АФ,attr,omitempty"`
	UT  string `xml:"xmlns:УТ,attr,omitempty"`
	VZL string `xml:"xmlns:ВЗЛ,attr,omitempty"`
}
