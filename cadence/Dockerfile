# syntax=docker/dockerfile:1
FROM golang:1.18.1-bullseye

WORKDIR /cadence
COPY ./* ./
RUN go mod download

# create a binary at root called 'cadence-server'
RUN go build -o /cadence-server

# TODO: ship config

EXPOSE 8080

RUN apt update
RUN apt install telnet -y
RUN apt install net-tools -y
RUN apt install iputils-ping -y
RUN apt install traceroute -y

# create a user named 'cadence', which owns the binary
RUN useradd -s /bin/bash cadence
RUN chown cadence /cadence-server

USER cadence
CMD [ "/cadence-server" ]
