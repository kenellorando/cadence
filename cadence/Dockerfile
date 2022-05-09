# syntax=docker/dockerfile:1
FROM golang:1.18.1-bullseye

WORKDIR /cadence
COPY ./* ./
RUN go mod download

# create a binary at root called 'cadence-server'
RUN go build -o /cadence-server

# TODO: ship config

EXPOSE 8080

# create a user named 'cadence', which owns the binary
RUN useradd -s /bin/bash cadence
RUN chown cadence /cadence-server

USER cadence
CMD [ "/cadence-server" ]
