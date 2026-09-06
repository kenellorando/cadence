# syntax=docker/dockerfile:1
ARG ARCH=
FROM ${ARCH}debian:trixie-slim
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"

# The base image tag is the pin. Debian ships exactly one liquidsoap version per
# suite -- 2.3.x on trixie -- so the suite already fixes the version, while
# leaving the package itself unpinned lets security rebuilds through. Pinning an
# exact revision on top of the base is what broke this image on bullseye: the
# revision aged out of the archive and the build started failing on 404s with no
# change on our side.
RUN apt-get update && \
    apt-get install -y --no-install-recommends liquidsoap && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 1234
USER liquidsoap
CMD [ "liquidsoap", "/etc/liquidsoap/cadence.liq" ]
