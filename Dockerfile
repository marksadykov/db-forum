FROM golang:1.15.2-buster AS build

MAINTAINER mark sadykov

RUN go get -t "github.com/fasthttp/router"

RUN go get -t "github.com/jackc/pgx"
#
RUN go get -t "github.com/jackc/pgx/pgxpool"
##
#RUN go get -t "github.com/jackc/tern/migrate"

RUN go get -t "github.com/labstack/gommon/log"

RUN go get -t "github.com/valyala/fasthttp"

#RUN mkdir /go/src/projectdb
#
#COPY main.go /go/src/projectdb
#
#WORKDIR /go/src/projectdb
#
#RUN go build -o db-forum main.go

ADD . /opt/app
WORKDIR /opt/app
RUN go build ./main.go

FROM ubuntu:20.04 AS release

MAINTAINER mark sadykov

RUN apt-get -y update && apt-get install -y locales gnupg2
RUN locale-gen en_US.UTF-8
RUN update-locale LANG=en_US.UTF-8

ENV PGVER 12
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update -y && apt-get install -y postgresql postgresql-contrib

ENV TZ=Russia/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

RUN mkdir /fordump

COPY mark.sql /fordump

USER postgres

RUN /etc/init.d/postgresql start &&\
    psql --command "CREATE USER root WITH SUPERUSER PASSWORD 'root';" &&\
#    psql --command "alter user root with password 'root';" &&\
#    createdb -U root db_forum_docker &&\
    psql --command "create database root owner root;" &&\
    psql --command "SET TIME ZONE 'Europe/Moscow';" &&\
#    PGPASSWORD="root" pg_restore -h localhost -U root -F t -d root /fordump/dumpfile &&\
    PGPASSWORD="root" psql -h localhost -d root -U root -p 5432 -a -q -f /fordump/mark.sql &&\
    /etc/init.d/postgresql stop

RUN echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

RUN echo "listen_addresses='*'" >> /etc/postgresql/$PGVER/main/postgresql.conf

VOLUME  ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

#USER root
#
#COPY --from=build /go/src/projectdb/db-forum /usr/bin/db-forum
#
#EXPOSE 5432
#EXPOSE 5000


EXPOSE 5432

USER root

WORKDIR /usr/src/app

COPY . .
COPY --from=build /opt/app/main .

EXPOSE 5000

#CMD service postgresql start && db-forum
CMD service postgresql start && ./main