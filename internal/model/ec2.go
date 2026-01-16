package model

type EC2Instance struct {
	InstanceID   string `header:"ID"`
	Name         string `header:"Name"`
	State        string `header:"State"`
	InstanceType string `header:"Type"`
	AZ           string `header:"AZ"`
	PrivateIP    string `header:"Private IP"`
	PublicIP     string `header:"Public IP"`
	LaunchTime   string `header:"Launch Time"`
}
