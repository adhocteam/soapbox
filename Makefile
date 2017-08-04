SOAPBOX_PKGS := $(shell go list ./... | grep -v /vendor/)

LDFLAGS :=

GIT_COMMIT = $(shell git rev-parse --short HEAD)

LDFLAGS += -X github.com/adhocteam/soapbox/version.GitCommit=${GIT_COMMIT}
LDFLAGS += -X "github.com/adhocteam/soapbox/version.BuildTime=$(shell date)"

all:
	go install -ldflags '$(LDFLAGS)' $(SOAPBOX_PKGS)

protobufs:
	make -C soapboxpb
	make -C web

models:
	PGSSLMODE=disable xo pgsql://localhost/soapbox_dev -o models

.PHONY: models

server:
	AWS_REGION=us-east-1 PGSSLMODE=disable PGDATABASE=soapbox_dev soapboxd
