# syntax=docker/dockerfile:1
ARG ARCH=
FROM ${ARCH}golang:1.20.0-bullseye
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
WORKDIR /cadence/server
COPY ./* ./
RUN go mod download
RUN go build -o /cadence/cadence-server

RUN useradd -s /bin/bash cadence
RUN touch /cadence/music-metadata.db
RUN chown cadence:cadence /cadence/ /cadence/* /cadence/cadence-server
RUN chmod u+wrx /cadence/ /cadence/* 

EXPOSE 8080
USER cadence
CMD [ "/cadence/cadence-server" ]
