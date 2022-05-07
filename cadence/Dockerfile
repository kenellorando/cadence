# syntax=docker/dockerfile:1
FROM golang:1.18.1-bullseye

WORKDIR /build
COPY ./* ./
RUN go mod download
RUN go build -o /cadence

# TODO: ship config

EXPOSE 8080
RUN useradd -s /bin/bash cadence
RUN chown cadence /cadence

USER cadence
CMD [ "/cadence" ]
