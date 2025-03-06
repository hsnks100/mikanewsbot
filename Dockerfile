# ubuntu
#FROM ubuntu:latest
FROM golang:1.21-bullseye
RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y build-essential git pkg-config libunistring-dev libaom-dev libdav1d-dev bzip2 nasm wget yasm ca-certificates
ENV PATH="/usr/local/go/bin:${PATH}"
COPY ./ /app
WORKDIR /app
RUN ls .
RUN go mod download
RUN go build -o /app/bin/mikanewsbot
RUN cp config.yaml /app/bin/config.toml
RUN touch .env
RUN cp .env /app/bin/.env

WORKDIR /app/bin
ENTRYPOINT ["/app/bin/mikanewsbot"]
