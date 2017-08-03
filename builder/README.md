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

If the `RELEASE` variable is set with a committish, the build process will create and upload a second `.tar.gz` file with the comittish included in the filename and upload said file to S3.

```
$ APPLICATION_SLUG=example-web-app \
    PATH_TO_CODE=/path/to/example-web-app \
    RELEASE=12469f467db5b8cc2c21ce346da108015823f87c \
    ./build.sh
```
