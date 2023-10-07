# syntax=docker/dockerfile:1
## Build API
FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.21-bullseye as api-builder
ARG TARGETPLATFORM BUILDPLATFORM TARGETOS TARGETARCH
WORKDIR /api-builder
COPY ./* ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o ./cadence-api

## Build UI
FROM node:20-alpine3.17 as ui-builder
WORKDIR /ui-builder
COPY ./ui/src ./src
COPY ./ui/static ./static
COPY ./ui/package*.json ./
COPY ./ui/*.config.js ./
RUN npm install
## Produces a static production "build" directory, see svelte.config.js
RUN npm run build

ARG ARCH=
FROM ${ARCH}golang:1.21-alpine
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
## The API server serves public/ as a static frontend site.
COPY --from=ui-builder /ui-builder/build /cadence/api/public
COPY --from=api-builder /api-builder/cadence-api /cadence/cadence-api

RUN adduser --disabled-password --gecos "" cadence
RUN chown cadence /cadence/ /cadence/* /cadence/cadence-api
RUN chmod u+wrx /cadence/ /cadence/* 

EXPOSE 8080
USER cadence
CMD [ "/cadence/cadence-api" ]
