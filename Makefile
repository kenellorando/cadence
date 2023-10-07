build-init:
	docker buildx create --platform linux/arm/v7,linux/amd64 --use --name multiarch

build-cadence:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag "kenellorando/cadence:latest" --tag kenellorando/cadence:$(VERSION) --file ./src/cadence.Dockerfile ./src/

build-cadence_icecast2:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_icecast2:latest --tag kenellorando/cadence_icecast2:$(VERSION) --file ./src/icecast2.Dockerfile ./src/

build-cadence_liquidsoap:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_liquidsoap:latest --tag kenellorando/cadence_liquidsoap:$(VERSION) --file ./src/liquidsoap.Dockerfile ./src/

build-all: build-cadence_api build-cadence_icecast2 build-cadence_liquidsoap
