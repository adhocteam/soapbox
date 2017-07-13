protobufdir = ../soapboxpb
protobufs = $(wildcard $(protobufdir)/*.proto)

all:
	bundle exec grpc_tools_ruby_protoc -I $(protobufdir) --ruby_out=lib --grpc_out=lib $(protobufs)
