# syntax=docker/dockerfile:1
ARG ARCH=
FROM ${ARCH}debian:11.3-slim
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"

RUN ln -s /usr/bin/dpkg-split /usr/sbin/dpkg-split && \
    ln -s /usr/bin/dpkg-deb /usr/sbin/dpkg-deb && \
    ln -s /bin/rm /usr/sbin/rm && \
    ln -s /bin/tar /usr/sbin/tar

RUN apt update
RUN apt install liquidsoap -y
RUN apt clean
EXPOSE 1234
USER liquidsoap
CMD [ "liquidsoap", "-t", "/etc/liquidsoap/cadence.liq" ]
