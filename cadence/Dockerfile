# syntax=docker/dockerfile:1
FROM golang:1.18.1-bullseye
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
WORKDIR /cadence
COPY ./* ./
RUN go mod download
RUN go build -o /cadence-server
RUN useradd -s /bin/bash cadence
RUN chown cadence /cadence-server
EXPOSE 8080
USER cadence
CMD [ "/cadence-server" ]
