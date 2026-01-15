package model

const (
	NodeRoute53     = "Route53"
	NodeCloudFront  = "CloudFront"
	NodeALB         = "ALB"
	NodeNLB         = "NLB"
	NodeCLB         = "CLB"
	NodeTargetGroup = "TargetGroup"
	NodeOrigin      = "Origin"
	NodeDNS         = "DNS"
)

type TraceNode struct {
	Type     string
	Name     string
	ID       string
	Value    string
	Status   string
	Children []TraceNode
}

type TraceResult struct {
	Domain string
	Hops   []TraceNode
}
