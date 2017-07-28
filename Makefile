SOAPBOX_PKGS := $(shell go list ./... | grep -v /vendor/)

all:
	go install $(SOAPBOX_PKGS)

protobufs:
	go generate $(SOAPBOX_PKGS)
	make -C web

models:
	PGSSLMODE=disable xo pgsql://localhost/soapbox_dev -o models
	-rm models/environment.xo.go

.PHONY: models

server:
	AWS_REGION=us-east-1 PGSSLMODE=disable PGDATABASE=soapbox_dev soapboxd
