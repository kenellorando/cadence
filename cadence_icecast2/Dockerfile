# syntax=docker/dockerfile:1
ARG ARCH=
FROM ${ARCH}alpine:3.14
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
RUN apk update && apk add icecast
EXPOSE 8000
USER icecast
CMD [ "icecast", "-c", "/etc/icecast/cadence.xml" ]
