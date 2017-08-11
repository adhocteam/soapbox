FROM golang:1.8
RUN mkdir -p /go/src/github.com/adhocteam/soapbox
COPY . /go/src/github.com/adhocteam/soapbox
WORKDIR /go/src/github.com/adhocteam/soapbox
ENV CGO_ENABLED=0
RUN make all

FROM alpine:latest
MAINTAINER ops@adhocteam.us
RUN apk update && apk add ca-certificates docker terraform
WORKDIR /root/
COPY --from=0 /go/bin/soapboxd .
CMD ["./soapboxd"]