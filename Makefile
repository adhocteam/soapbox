SOAPBOX_PKGS := $(shell go list ./... | grep -v /vendor/)

LDFLAGS :=

VERSION = $(shell cat VERSION)
GIT_COMMIT = $(shell git rev-parse --short HEAD)

LDFLAGS += -X github.com/adhocteam/soapbox/buildinfo.Version=${VERSION}
LDFLAGS += -X github.com/adhocteam/soapbox/buildinfo.GitCommit=${GIT_COMMIT}
LDFLAGS += -X "github.com/adhocteam/soapbox/buildinfo.BuildTime=$(shell date)"

all:
	go install -ldflags '$(LDFLAGS)' $(SOAPBOX_PKGS)

protobufs:
	make -C soapboxpb
	make -C web

models:
	PGSSLMODE=disable xo pgsql://soapbox@localhost:54320/soapbox_dev -o models --template-path models/templates/

.PHONY: models

server:
	AWS_REGION=us-east-1 PGSSLMODE=disable PGDATABASE=soapbox_dev soapboxd

docker-image:
	docker build -t soapbox/soapbox:$(GIT_COMMIT) .
