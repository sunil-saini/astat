package model

type LambdaFunction struct {
	Name         string `header:"Name"`
	Runtime      string `header:"Runtime"`
	LastModified string `header:"Last Modified"`
	Memory       string `header:"Memory (MB)"`
	Timeout      string `header:"Timeout (s)"`
}
