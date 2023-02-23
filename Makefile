build-init:
	docker buildx create --platform linux/arm/v7,linux/amd64 --use --name multiarch

build-cadence:
	docker buildx build --push --platform linux/arm/v7,linux/amd64 --tag kenellorando/cadence:latest --tag kenellorando/cadence:$(VERSION) --file ./cadence/Dockerfile.multiarch ./cadence/
