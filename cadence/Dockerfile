# syntax=docker/dockerfile:1
FROM golang:1.18.1-bullseye
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
WORKDIR /cadence
COPY ./* ./
RUN go mod download
RUN go build -o /cadence-server --tags "linux sqlite_fts5"
RUN useradd -s /bin/bash cadence
RUN chown cadence /cadence-server

RUN touch /cadence/music-metadata.db
RUN chown cadence:cadence /cadence/
RUN chown cadence:cadence /cadence/*
RUN chmod u+wrx /cadence/*
RUN chmod u+wrx /cadence/

EXPOSE 8080
USER cadence
CMD [ "/cadence-server" ]
