package model

type EC2Instance struct {
	InstanceID   string
	Name         string
	State        string
	InstanceType string
	AZ           string
	PrivateIP    string
	PublicIP     string
	LaunchTime   string
}
