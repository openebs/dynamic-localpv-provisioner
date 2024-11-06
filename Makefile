GO111MODULE ?= on
export GO111MODULE

# Determine the arch/os
ifeq (${XC_OS}, )
  XC_OS:=$(shell go env GOOS)
endif
export XC_OS

ifeq (${XC_ARCH}, )
  XC_ARCH:=$(shell go env GOARCH)
endif
export XC_ARCH

ARCH:=${XC_OS}_${XC_ARCH}
export ARCH


# list only the source code directories
PACKAGES = $(shell go list ./... | grep -v '/pkg/version\|tests')

# list only the integration tests code directories
PACKAGES_IT = $(shell go list ./... | grep -v 'pkg/client/generated' | grep 'tests')

# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in 
# the `buildscripts/push` script.
#
# The images of a project or company can then be grouped
# or hosted under a unique organization key like `openebs`
#
# Each component (container) will be pushed to a unique 
# repository under an organization. 
# Putting all this together, an unique uri for a given 
# image comprises of:
#   <registry url>/<image org>/<image repo>:<image-tag>
#
# IMAGE_ORG can be used to customize the organization 
# under which images should be pushed. 
# By default the organization name is `openebs`. 

ifeq (${IMAGE_ORG}, )
  IMAGE_ORG = openebs
endif

# If IMAGE_TAG is mentioned then TAG will be set to IMAGE_TAG
# If RELEASE_TAG is mentioned then TAG will be set to RELEAE_TAG
# If both are mentioned then TAG will be set to RELEASE_TAG

# Set the path to the Chart.yaml file
ROOT_DIR:=$(dir $(realpath $(firstword $(MAKEFILE_LIST))))
CHART_YAML:=${ROOT_DIR}/deploy/helm/charts/Chart.yaml
TAG := $(shell awk -F': ' '/^version:/ {print $$2}' $(CHART_YAML))

ifneq (${IMAGE_TAG}, )
  TAG=${IMAGE_TAG:v%=%}
endif

ifeq (${RELEASE_TAG}, )
  IMAGE_TAG=${RELEASE_TAG:v%=%}
endif

# Specify the name for the binaries
PROVISIONER_LOCALPV=provisioner-localpv

# Specify the name of the image
PROVISIONER_LOCALPV_IMAGE?=provisioner-localpv

# Final variable with image org, name and tag
PROVISIONER_LOCALPV_IMAGE_TAG=${IMAGE_ORG}/${PROVISIONER_LOCALPV_IMAGE}:${TAG}

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/dynamic-localpv-provisioner"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

# Specify the kubeconfig path to a Kubernetes cluster 
# to run Hostpath integration tests
ifeq (${KUBECONFIG}, )
  KUBECONFIG=${HOME}/.kube/config
  export KUBECONFIG
endif

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg BRANCH=${BRANCH} --build-arg RELEASE_TAG=${RELEASE_TAG}

.PHONY: all
all: test provisioner-localpv-image

.PHONY: deps
deps:
	@echo "--> Tidying up submodules"
	@go mod tidy
	@echo "--> Veryfying submodules"
	@go mod verify


.PHONY: verify-deps
verify-deps: deps
	@if !(git diff --quiet HEAD -- go.sum go.mod); then \
		echo "go module files are out of date, please commit the changes to go.mod and go.sum"; exit 1; \
	fi

.PHONY: clean
clean: 
	go clean -testcache
	rm -rf bin

.PHONY: test
test: format vet
	@echo "--> Running go test";
	$(PWD)/buildscripts/test.sh ${XC_ARCH}

.PHONY: testv
testv: format
	@echo "--> Running go test verbose" ;
	@go test -v $(PACKAGES)

# Requires KUBECONFIG env and Ginkgo binary
.PHONY: integration-test
integration-test:
	@cd tests && sudo -E env "PATH=${PATH}" ginkgo -v --fail-fast -coverprofile="integration_coverage.txt" -covermode=atomic; 

.PHONY: format
format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES) $(PACKAGES_IT)

# -composite: avoid "literal copies lock value from fakePtr"
.PHONY: vet
vet:
	@echo "--> Running go vet"
	@go list ./... | xargs go vet -composites

.PHONY: verify-src
verify-src: 
	@echo "--> Checking for git changes post running tests";
	$(PWD)/buildscripts/check-diff.sh "format"


#Use this to build provisioner-localpv
.PHONY: provisioner-localpv
provisioner-localpv:
	@echo "----------------------------"
	@echo "--> provisioner-localpv    "
	@echo "----------------------------"
	@PNAME=${PROVISIONER_LOCALPV} CTLNAME=${PROVISIONER_LOCALPV} sh -c "'$(PWD)/buildscripts/build.sh'"

.PHONY: provisioner-localpv-image
provisioner-localpv-image: provisioner-localpv
	@echo "-------------------------------"
	@echo "--> provisioner-localpv image "
	@echo "-------------------------------"
	@cp bin/provisioner-localpv/${PROVISIONER_LOCALPV} buildscripts/provisioner-localpv/
	@cd buildscripts/provisioner-localpv && docker build -t ${PROVISIONER_LOCALPV_IMAGE_TAG} ${DBUILD_ARGS} . --no-cache
	@rm buildscripts/provisioner-localpv/${PROVISIONER_LOCALPV}


.PHONY: push
push:
	DIMAGE=${IMAGE_ORG}/${PROVISIONER_LOCALPV_IMAGE} ./buildscripts/push.sh

# include the buildx recipes
include Makefile.buildx.mk
