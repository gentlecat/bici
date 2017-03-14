FROM golang:latest

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
                    libpq-dev \
                    git \
                    ruby-full \
                    runit && \
    rm -rf /var/lib/apt/lists/*

# PostgreSQL client
RUN apt-key adv --keyserver ha.pool.sks-keyservers.net --recv-keys B97B0AFCAA1A47F044F244A07FCC7D46ACCC4CF8
ENV PG_MAJOR 9.6
RUN echo 'deb http://apt.postgresql.org/pub/repos/apt/ jessie-pgdg main' $PG_MAJOR > /etc/apt/sources.list.d/pgdg.list
RUN apt-get update \
    && apt-get install -y postgresql-client-$PG_MAJOR \
    && rm -rf /var/lib/apt/lists/*
# Specifying password so that client doesn't ask scripts for it...
ENV PGPASSWORD "summits"

ENV GOPATH /go
RUN mkdir -p $GOPATH
WORKDIR /go

RUN go get -u github.com/peterbourgon/runsvinit
RUN cp $GOPATH/bin/runsvinit /usr/local/bin/

COPY . /go/src/go.roman.zone/bici
RUN go get -v ./...
RUN go build go.roman.zone/bici
COPY ./res /go/bin/res

# Styling
WORKDIR /go/bin/res/static/styles
RUN gem install sass
RUN scss main.scss:main.css

# Service
COPY ./web.service /etc/sv/web/run
RUN chmod 755 /etc/sv/web/run && \
    ln -sf /etc/sv/web /etc/service/

WORKDIR /go
EXPOSE 80
ENTRYPOINT ["/usr/local/bin/runsvinit"]
