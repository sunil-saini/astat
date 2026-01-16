package model

type Route53HostedZone struct {
	ID      string `header:"ID"`
	Name    string `header:"Name"`
	Type    string `header:"Type"`
	Records string `header:"Records"`
}

type Route53Record struct {
	ZoneName string `header:"Zone"`
	Name     string `header:"Name"`
	Type     string `header:"Type"`
	TTL      string `header:"TTL"`
	Value    string `header:"Value"`
}
