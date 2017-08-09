# Soapbox -- the web application hosting platform

Soapbox provides managed web application hosting services, encapsulating best-practices for deployment and operations. Soapbox is a platform on top of cloud services that allows teams to focus on development and provides scaling and monitoring without custom configuration.

## Getting started

### Requirements

 - Go 1.8 or greater - see directions [here](https://golang.org/doc/install)
 - Ruby 2.2 or greater - see directions [here](https://www.ruby-lang.org/en/documentation/installation/)
 - PostgreSQL 9.5 or greater - [Ubuntu](https://www.digitalocean.com/community/tutorials/how-to-install-and-use-postgresql-on-ubuntu-16-04) | [Mac](https://solidfoundationwebdev.com/blog/posts/how-to-install-postgresql-using-brew-on-osx)
 - Terraform 0.9.11 or greater - Download and install Terraform from [here](https://www.terraform.io/downloads.html).

### Running the server and client locally

1. **Create a PostgreSQL db and its initial schema:**
``` shell
$ createdb soapbox_dev
$ psql -f db/schema.sql -d soapbox_dev
```

2. **Build and install the Soapbox API server (soapboxd) and CLI client (soapboxcli):**
``` shell
$ mkdir -p $(go env GOPATH)/src/github.com/adhocteam
$ go get github.com/adhocteam/soapbox/...
$ cd $(go env GOPATH)/src/github.com/adhocteam/soapbox
$ go get -u github.com/golang/dep/cmd/dep
$ go get -d -v ./...
$ go install ./...
```

3. **Run the API server:**
``` shell
$ PGDATABASE=soapbox_dev PGSSLMODE=disable soapboxd &
```
* If your database user is password protected, you may need to pass `PGPASSWORD=yourpgpass` to the command above as well.

4. **Try out the CLI client:**
``` shell
$ soapboxcli list-applications
```

5. **Install the web client:**
``` shell
$ cd web
$ gem install bundler
$ bundle install
```

6. **Run the web client and try it out:**
``` shell
$ bin/rails server &
$ open http://localhost:3000/
```

### GitHub OAuth

In order to use Soapbox with private repositories, you must grant the
application access through OAuth.
1. Go to the [OAuth applications](https://github.com/settings/developers)
page and click on `Register a new application`.
2. Give your application a name and a homepage URL (this could be
`localhost:3000` or a registered domain).
3. For `Authorization callback URL`, enter your homepage URL followed by
`/auth/github/callback`.
4. When you submit, you will see a `Client ID` and a `Client Secret`.
Set these as the environment variables `GITHUB_OAUTH_CLIENT_ID` and
`GITHUB_OAUTH_CLIENT_SECRET` (be sure to restart your Rails server after
these are set).
5. Create a user in the Soapbox web UI, click `Link to GitHub` on your
profile page, and grant the requested permissions.

## Developing Soapbox

### Making changes to protobufs

Soapbox
uses
[Protocol Buffers](https://developers.google.com/protocol-buffers/)
via [gRPC](https://grpc.io/) for clients and servers to exchange
messages and call API methods. These definitions are stored in the
`soapboxpb` directory in `.proto` files. If you change these files,
you must re-generate the Go and Ruby code that the API server and the
Rails app rely on, respectively.

``` shell
# Go code
$ go generate ./...
$ go install ./...
# Ruby code
$ make -C web
```

### Go dependencies

Soapbox uses [dep](https://github.com/golang/dep) for dependency management. Follow the below flow to add imports:

- Add the import to the code
- Run `dep ensure` to make sure that the manifest, lock file, and vendor folder are updated

Running these steps will clone the repo under the vendor directory, and remembers the revision used so that everyone who works on the project is guaranteed to be using the same version of dependencies.

## Design documentation

 * [Architecture document](https://docs.google.com/document/d/1hArh6EGNfa23O1mPKVeq_OjfA4AiCBEvc-k07xsb4t4/edit#)
 * [Product announcement](https://docs.google.com/document/d/1njbQ0hTEHrA8kYHe-_N_0K-Z6lcyFU-taSI13bQPDPo/edit#heading=h.fcmb7lh1usjg)
