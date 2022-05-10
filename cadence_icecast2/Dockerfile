# syntax=docker/dockerfile:1
FROM debian:11.3-slim
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
RUN apt update
RUN apt install icecast2 -y
RUN apt clean
EXPOSE 8000
RUN useradd -g icecast icecast
RUN chown -R icecast:icecast /var/log/icecast2/
USER icecast
CMD [ "icecast2", "-c", "/etc/icecast2/cadence.xml" ]
