package model

type S3Bucket struct {
	Name         string `header:"Name"`
	Region       string `header:"Region"`
	CreationDate string `header:"Creation Date"`
}
