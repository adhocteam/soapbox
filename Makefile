all:
	go install ./...

protobufs:
	go generate ./...
	make -C web

models:
	PGSSLMODE=disable xo pgsql://localhost/soapbox_dev -o models

.PHONY: models

server:
	AWS_REGION=us-east-1 PGSSLMODE=disable PGDATABASE=soapbox_dev soapboxd
