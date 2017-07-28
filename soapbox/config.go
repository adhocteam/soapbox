package soapbox

type Config struct {
	AmiName      string
	Domain       string
	IamProfile   string // IAM instance profile
	KeyName      string
	InstanceType string
}
