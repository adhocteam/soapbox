# Soapbox AMI

The Soapbox AMI uses the plain AWS Linux AMI as its base.

The `soapbox-app.json` packer template uses `playbooks/soapbox-app.yml` to provision an AMI with:

- Runit
- Nginx
- Docker

It installs a simple `nginx.conf` that proxies to port `9090`.

As such, it expects `docker` to publish an application's container port `9090`.

## Usage

To build a new AMI:

```
$ cd ops/aws/ami
$ packer build soapbox-app.json
```
