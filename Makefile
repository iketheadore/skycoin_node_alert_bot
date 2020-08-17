.DEFAULT_GOAL := build
.PHONY: build push

IMG = registry.skycoin.com/skycoin-node-alert-bot

build: ## build docker image
	docker build -t $(IMG) .

push: ## push image to registry
	docker push $(IMG)
