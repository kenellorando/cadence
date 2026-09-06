# syntax=docker/dockerfile:1
ARG ARCH=
# Pin the base to a specific Alpine release. A floating alpine:3 combined with
# an exact package revision below cannot both hold: the revision ages out of the
# index and the build breaks with no change on our side.
FROM ${ARCH}alpine:3.22
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
RUN apk update && apk add icecast=2.4.4-r13
EXPOSE 8000
USER icecast
CMD [ "icecast", "-c", "/etc/icecast/cadence.xml" ]
