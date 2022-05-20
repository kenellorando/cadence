build-init:
	docker buildx create --use

build-cadence:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence:latest --file ./cadence/Dockerfile.buildx ./cadence/

build-cadence_icecast2:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_icecast2:latest --file ./cadence_icecast2/Dockerfile.buildx ./cadence_icecast2/

build-cadence_liquidsoap:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_liquidsoap:latest --file ./cadence_liquidsoap/Dockerfile.buildx ./cadence_liquidsoap/

build-all: build-init build-cadence build-cadence_icecast2 build-cadence_liquidsoap
