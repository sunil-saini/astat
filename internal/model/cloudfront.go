package model

type CloudFrontDistribution struct {
	ID            string
	Domain        string
	Status        string
	Aliases       string
	LastUpdated   string
	Origins       map[string]string // ID -> DomainName
	DefaultOrigin string
	Behaviors     []CloudFrontBehavior
}

type CloudFrontBehavior struct {
	PathPattern    string
	TargetOriginID string
}
