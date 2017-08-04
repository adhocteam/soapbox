DEST = ../proto
PROTOBUFS = $(wildcard *.proto)
PROTOC_FLAGS  = -I .

all:
	protoc $(PROTOC_FLAGS) --go_out=plugins=grpc:$(DEST) $(PROTOBUFS)
