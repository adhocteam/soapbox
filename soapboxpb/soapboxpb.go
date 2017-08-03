package soapboxpb

//go:generate protoc --proto_path=. --go_out=plugins=grpc:. soapbox.proto application.proto deployment.proto environment.proto version.proto
