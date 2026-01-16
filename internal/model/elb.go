package model

type LoadBalancer struct {
	Type      string `header:"Type"`
	Name      string `header:"Name"`
	Scheme    string `header:"Scheme"`
	CreatedAt string `header:"Created At"`
	DNSName   string `header:"DNS"`
	ARN       string // For v2
}

type Listener struct {
	ARN            string
	Protocol       string
	Port           int32
	DefaultActions []Action
}

type Rule struct {
	ARN        string
	Priority   string
	IsDefault  bool
	Conditions []Condition
	Actions    []Action
}

type Condition struct {
	Field  string // "host-header" or "path-pattern"
	Values []string
}

type Action struct {
	Type           string
	TargetGroupARN string
}

type TargetGroup struct {
	Name            string
	Protocol        string
	Port            int32
	TargetType      string
	LoadBalancerARN string
	ARN             string
}

type InstanceHealth struct {
	InstanceID string
	Name       string
	State      string
	Reason     string
}
