package model

type RDSCluster struct {
	ClusterIdentifier string
	Status            string
	Engine            string
	EngineVersion     string
	MultiAZ           string
	IsPublic          string
	InstanceCount     int
	StorageType       string
	CreateTime        string
}

type RDSInstance struct {
	ClusterIdentifier  string
	InstanceIdentifier string
	Role               string
	Engine             string
	EngineVersion      string
	DBInstanceStatus   string
	Endpoint           string
	InstanceClass      string
	AvailabilityZone   string
}
