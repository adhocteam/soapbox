FROM alpine:latest
RUN apk add --no-cache ca-certificates docker
COPY build-soapbox-app.sh /soapbox/
CMD /soapbox/build-soapbox-app.sh