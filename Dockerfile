FROM golang:1.8-alpine as builder
RUN apk add --update --no-cache build-base git
RUN mkdir -p /go/src/github.com/adhocteam/soapbox
COPY . /go/src/github.com/adhocteam/soapbox
WORKDIR /go/src/github.com/adhocteam/soapbox
ENV CGO_ENABLED=0
RUN make all

FROM alpine:latest
MAINTAINER ops@adhocteam.us
RUN apk add --update --no-cache ca-certificates docker git terraform
WORKDIR /root/
COPY --from=builder /go/bin/soapboxd .
CMD ["./soapboxd"]
