package model

type Route53HostedZone struct {
	ID      string
	Name    string
	Type    string
	Records string
}

type Route53Record struct {
	ZoneName string
	Name     string
	Type     string
	TTL      string
	Value    string
}
