# syntax=docker/dockerfile:1
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.21-bullseye as api-builder
ARG TARGETPLATFORM BUILDPLATFORM TARGETOS TARGETARCH
WORKDIR /api-builder
COPY ./* ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o ./cadence-api

ARG ARCH=
FROM ${ARCH}golang:1.21-alpine
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
COPY --from=api-builder /api-builder/cadence-api /cadence/cadence-api

RUN adduser --disabled-password --gecos "" cadence
RUN chown cadence /cadence/ /cadence/* /cadence/cadence-api
RUN chmod u+wrx /cadence/ /cadence/* 

EXPOSE 8080
USER cadence
CMD [ "/cadence/cadence-api" ]
