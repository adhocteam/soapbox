package soapbox

type Config struct {
	AmiId        string
	Domain       string
	IamProfile   string // IAM instance profile
	KeyName      string
	InstanceType string
}
