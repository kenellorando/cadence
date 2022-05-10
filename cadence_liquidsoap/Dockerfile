# syntax=docker/dockerfile:1
FROM debian:11.3-slim
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
RUN apt update
RUN apt install liquidsoap -y
RUN apt clean
EXPOSE 1234
USER liquidsoap
CMD [ "liquidsoap", "-t", "/etc/liquidsoap/cadence.liq" ]
