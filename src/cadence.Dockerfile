# syntax=docker/dockerfile:1
# The UI is a SvelteKit app prerendered to static files. It builds on the native
# platform because the output is plain HTML/CSS/JS with nothing arch-specific.
FROM --platform=${BUILDPLATFORM:-linux/amd64} node:22-alpine AS ui
WORKDIR /ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

FROM --platform=${BUILDPLATFORM:-linux/amd64} golang:1.26-bookworm as builder
ARG TARGETPLATFORM BUILDPLATFORM TARGETOS TARGETARCH
WORKDIR /cadence
COPY server/ ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags="-w -s" -o /cadence-server

ARG ARCH=
# The build above is CGO_ENABLED=0, so the binary is static and the runtime
# needs nothing from the Go toolchain. Shipping golang:*-alpine here cost
# roughly 260MB of compiler to run a 10MB binary.
FROM ${ARCH}alpine:3.24
LABEL maintainer="Ken Ellorando (kenellorando.com)"
LABEL source="github.com/kenellorando/cadence"
COPY --from=ui /ui/build /cadence/server/public
COPY --from=builder /cadence-server /cadence/cadence-server

RUN adduser --disabled-password --gecos "" cadence
RUN chown cadence /cadence/ /cadence/* /cadence/cadence-server
RUN chmod u+wrx /cadence/ /cadence/*

EXPOSE 8080
USER cadence
CMD [ "/cadence/cadence-server" ]
