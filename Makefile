build-init:
	docker buildx create --platform linux/arm/v7,linux/amd64 --use --name multiarch

build-cadence:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence:latest --file ./cadence/Dockerfile.multiarch ./cadence/

build-cadence_icecast2:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_icecast2:latest ./cadence_icecast2/

build-cadence_liquidsoap:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence_liquidsoap:latest ./cadence_liquidsoap/

build-all: build-init build-cadence build-cadence_icecast2 build-cadence_liquidsoap
