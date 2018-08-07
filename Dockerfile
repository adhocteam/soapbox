FROM golang:1.9-alpine as builder
RUN apk add --update --no-cache build-base git
RUN mkdir -p /go/src/github.com/adhocteam/soapbox
COPY . /go/src/github.com/adhocteam/soapbox
WORKDIR /go/src/github.com/adhocteam/soapbox
ENV CGO_ENABLED=0
RUN make all

FROM alpine:latest
MAINTAINER ops@adhocteam.us

ENV TERRAFORM_VERSION=0.10.5
ENV TERRAFORM_SHA256SUM=acec7133ffa00da385ca97ab015b281c6e90e99a41076ede7025a4c78425e09f

RUN apk add --update --no-cache ca-certificates curl docker git openssh && \
    curl https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip > terraform_${TERRAFORM_VERSION}_linux_amd64.zip && \
    echo "${TERRAFORM_SHA256SUM}  terraform_${TERRAFORM_VERSION}_linux_amd64.zip" > terraform_${TERRAFORM_VERSION}_SHA256SUMS && \
    sha256sum -cs terraform_${TERRAFORM_VERSION}_SHA256SUMS && \
    unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /bin && \
    rm -f terraform_${TERRAFORM_VERSION}_linux_amd64.zip

WORKDIR /root/
RUN mkdir ops
COPY --from=builder /go/bin/soapboxd .
COPY ./ops ./ops
COPY ./templates ./templates
CMD ["./soapboxd"]
