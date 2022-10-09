build-init:
	docker buildx create --platform linux/arm/v7,linux/amd64 --use --name multiarch

build-cadence:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence:latest --tag kenellorando/cadence:$(VERSION) --file ./cadence/Dockerfile.multiarch ./cadence/

build-cadence_icecast2:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_icecast2:latest --tag kenellorando/cadence_icecast2:$(VERSION) --file ./cadence_icecast2/Dockerfile ./cadence_icecast2/

build-cadence_liquidsoap:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_liquidsoap:latest --tag kenellorando/cadence_liquidsoap:$(VERSION) --file ./cadence_liquidsoap/Dockerfile ./cadence_liquidsoap/

build-all: build-cadence build-cadence_icecast2 build-cadence_liquidsoap
