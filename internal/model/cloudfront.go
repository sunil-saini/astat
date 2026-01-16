package model

type CloudFrontDistribution struct {
	ID            string            `header:"ID"`
	Domain        string            `header:"Domain"`
	Status        string            `header:"Status"`
	Aliases       string            `header:"Aliases"`
	LastUpdated   string            `header:"LastUpdated"`
	Origins       map[string]string // ID -> DomainName
	DefaultOrigin string
	Behaviors     []CloudFrontBehavior
}

type CloudFrontBehavior struct {
	PathPattern    string
	TargetOriginID string
}
