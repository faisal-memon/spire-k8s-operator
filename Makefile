.PHONY: default build all

default: build

all: build lint

############################################################################
# Vars
############################################################################

go_version := $(shell cat .go-version)
DOCKER_REGISTRY?=gcr.io/nginx-mesh-dev

#############################################################################
# Build Targets
#############################################################################

.PHONY: build

build:
	go build -o build/bin/spire-k8s-operator ./cmd/manager

#############################################################################
# Docker Image
#############################################################################

.PHONY: image
image: build/Dockerfile 
	docker build --build-arg goversion=$(go_version) -t spire-k8s-operator -f build/Dockerfile .
	docker tag spire-k8s-operator:latest $(DOCKER_REGISTRY)/spire-k8s-operator:latest
	docker push $(DOCKER_REGISTRY)/spire-k8s-operator:latest
	sed 's|$$(DOCKER_REGISTRY)|$(DOCKER_REGISTRY)|g' deploy/operator.yaml.tmpl > deploy/operator.yaml
