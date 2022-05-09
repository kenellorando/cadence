# syntax=docker/dockerfile:1
FROM debian:stable
RUN apt update
RUN apt install liquidsoap -y

RUN apt install telnet -y
RUN apt install net-tools -y
RUN apt install iputils-ping -y

RUN apt clean

EXPOSE 1234
USER liquidsoap
CMD [ "liquidsoap", "-t", "/etc/liquidsoap/cadence.liq" ]
