package model

type RDSCluster struct {
	ClusterIdentifier string `header:"Identifier"`
	Status            string `header:"Status"`
	Engine            string `header:"Engine"`
	EngineVersion     string `header:"Engine Version"`
	MultiAZ           string `header:"Multi-AZ"`
	IsPublic          string `header:"Public Access"`
	InstanceCount     int    `header:"Instance Count"`
	StorageType       string `header:"Storage Type"`
	CreateTime        string `header:"Created At"`
}

type RDSInstance struct {
	ClusterIdentifier  string `header:"Cluster"`
	InstanceIdentifier string `header:"Identifier"`
	Role               string `header:"Role"`
	Engine             string `header:"Engine"`
	EngineVersion      string `header:"Engine Version"`
	DBInstanceStatus   string `header:"Status"`
	Endpoint           string `header:""`
	InstanceClass      string `header:"Class"`
	AvailabilityZone   string `header:"AZ"`
}
