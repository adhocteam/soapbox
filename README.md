# Soapbox -- the web application hosting platform

Soapbox provides managed web application hosting services, encapsulating best-practices for deployment and operations. Soapbox is a platform on top of cloud services that allows teams to focus on development and provides scaling and monitoring without custom configuration.

## Getting started

### Requirements

 - Go 1.8 or greater
 - Ruby 2.2 or greater
 - PostgreSQL 9.5 or greater

### Running the server and client locally

1. **Create a PostgreSQL db and its initial schema:**
``` shell
$ createdb soapbox_dev
$ psql -f db/schema.sql -d soapbox_dev
```

2. **Build and install the Soapbox API server (soapboxd) and CLI client (soapboxcli):**
``` shell
$ go install ./...
```

3. **Run the API server:**
``` shell
$ PGDATABASE=soapbox_dev PGSSLMODE=disable soapboxd &
```

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

## Design documentation

 * [Architecture document](https://docs.google.com/document/d/1hArh6EGNfa23O1mPKVeq_OjfA4AiCBEvc-k07xsb4t4/edit#)
 * [Product announcement](https://docs.google.com/document/d/1njbQ0hTEHrA8kYHe-_N_0K-Z6lcyFU-taSI13bQPDPo/edit#heading=h.fcmb7lh1usjg)
