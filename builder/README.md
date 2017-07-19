# soapbox builder

The builder is a Docker image & script that defines a container that can build soapbox applications and place the result in S3. By using containers for builds, soapbox isolates each application build in a well-known, secure build environment.

## Setup

You must build the soapbox image before using the `build.sh` script.

```
docker build -t soapbox .
```

## Usage

This is an example of how to build the [example-web-app](https://github.com/adhocteam/example-web-app) image.

The `build.sh` script will instantiate a new builder container and mount the host Docker process as a volume. The builder container invokes the `docker build` command, which builds the application container image, then invokes `docker save` to write the container image to a .tar file. The script then uploads the Gzipped .tar to S3.

```
$ APPLICATION_SLUG=example-web-app PATH_TO_CODE=/Users/cgansen/projects/ah/example-web-app ./build.sh
Sending build context to Docker daemon  9.217MB
Step 1/10 : FROM golang:1.8 as builder
1.8: Pulling from library/golang
9f0706ba7422: Pull complete
d3942a742d22: Pull complete
c6575234aef3: Pull complete
5b93ef2d289f: Pull complete
4519a0c26511: Pull complete
8c8468a2a816: Pull complete
163e6ebf1e88: Pull complete
Digest: sha256:b8bf21db6cf238db5f79cc3b2d773d0982367876d03dc5c8b8da5c336e85e612
Status: Downloaded newer image for golang:1.8
 ---> 6d9bf2aec386
Step 2/10 : MAINTAINER ops@adhocteam.us
 ---> Running in e2e9460f4b42
 ---> dc4223cbe89d
Removing intermediate container e2e9460f4b42
Step 3/10 : WORKDIR /go/src/app
 ---> 351ab61cef7d
Removing intermediate container 268eb63b71be
Step 4/10 : COPY . .
 ---> c164071091a1
Removing intermediate container c75b54ced77f
Step 5/10 : RUN CGO_ENABLED=0 go build -o app .
 ---> Running in 67fa7282d527
 ---> 54f3faef9f18
Removing intermediate container 67fa7282d527
Step 6/10 : FROM alpine:latest
 ---> 7328f6f8b418
Step 7/10 : MAINTAINER ops@adhocteam.us
 ---> Running in 41f1ef1afdd8
 ---> e74ba18a8c13
Removing intermediate container 41f1ef1afdd8
Step 8/10 : WORKDIR /root/
 ---> 77599ec81947
Removing intermediate container bcb243de63ce
Step 9/10 : COPY --from=0 /go/src/app/app .
 ---> eb9a89b6d72a
Removing intermediate container 8a259742069c
Step 10/10 : CMD ./app
 ---> Running in dc6f103303e4
 ---> eba945728e62
Removing intermediate container dc6f103303e4
Successfully built eba945728e62
Successfully tagged example-web-app:latest
upload: ../../example-web-app/example-web-app.tar.gz to s3://soapbox-app-images/example-web-app/example-web-app.tar.gz
```
