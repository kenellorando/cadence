# syntax=docker/dockerfile:1
FROM golang:1.18.1-bullseye as builder
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
WORKDIR /cadence/server
COPY ./* ./
RUN go mod download
RUN go build -o ./cadence-server --tags "linux sqlite_fts5"


FROM alpine:latest
WORKDIR /cadence/server
COPY --from=builder /cadence/server .
EXPOSE 8080
CMD [ "./cadence-server" ]
