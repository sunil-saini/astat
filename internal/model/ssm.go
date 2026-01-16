package model

type SSMParameter struct {
	Name         string `header:"Name"`
	Type         string `header:"Type"`
	LastModified string `header:"Last Modified"`
	ModifiedBy   string `header:"Modified By"`
}
