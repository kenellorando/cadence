# syntax=docker/dockerfile:1
ARG ARCH=
FROM ${ARCH}debian:11-slim
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"

RUN ln -s /usr/bin/dpkg-split /usr/sbin/dpkg-split && \
    ln -s /usr/bin/dpkg-deb /usr/sbin/dpkg-deb && \
    ln -s /bin/rm /usr/sbin/rm && \
    ln -s /bin/tar /usr/sbin/tar

RUN apt clean all
RUN apt update
RUN apt install liquidsoap=1.4.3-3 -y
RUN apt autoremove
EXPOSE 1234
USER liquidsoap
CMD [ "liquidsoap", "-t", "/etc/liquidsoap/cadence.liq" ]
