FROM golang:latest AS build

RUN mkdir /DB_TP

COPY . /DB_TP

WORKDIR /DB_TP

RUN go build -o db_tp app/app.go

FROM ubuntu:20.04 AS release

RUN apt-get update -y && apt-get install -y locales gnupg2
RUN locale-gen en_US.UTF-8
RUN update-locale LANG=en_US.UTF-8

ENV PGVER 12
ENV DEBIAN_FRONTEND noninteractive
RUN apt-get update -y && apt-get install -y postgresql postgresql-contrib

USER postgres

COPY init.sql /home

WORKDIR /home

RUN /etc/init.d/postgresql start &&\
    psql --command "ALTER USER postgres WITH PASSWORD 'postgres';" &&\
    createdb -E UTF8 forums &&\
    psql --command "\i '/home/init.sql'" &&\
    /etc/init.d/postgresql stop

RUN echo "listen_addresses='*'\n" >> /etc/postgresql/$PGVER/main/postgresql.conf
RUN echo "host all all 0.0.0.0/0 md5" >> /etc/postgresql/$PGVER/main/pg_hba.conf

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

COPY --from=build /DB_TP/db_tp /usr/bin/db_tp
COPY --from=build /DB_TP/config.json /home/config.json

EXPOSE 5000

CMD service postgresql start && db_tp config.json
