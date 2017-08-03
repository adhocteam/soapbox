package soapboxpb

//go:generate protoc --proto_path=. --go_out=plugins=grpc:../proto soapbox.proto application.proto deployment.proto environment.proto version.proto
