# Copyright © 2020 The OpenEBS Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
PACKAGES = $(shell go list ./... | grep -v 'vendor\|pkg/client/generated\|tests')

# list only the integration tests code directories
PACKAGES_IT = $(shell go list ./... | grep -v 'vendor\|pkg/client/generated' | grep 'tests')

ifeq (${IMAGE_TAG}, )
  IMAGE_TAG = ci
  export IMAGE_TAG
endif

ifeq (${TRAVIS_TAG}, )
  BASE_TAG = ci
  export BASE_TAG
else
  BASE_TAG = $(TRAVIS_TAG:v%=%)
  export BASE_TAG
endif

# The images can be pushed to any docker/image registeries
# like docker hub, quay. The registries are specified in 
# the `build/push` script.
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
  export IMAGE_ORG
endif

# Specify the date of build
DBUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Specify the docker arg for repository url
ifeq (${DBUILD_REPO_URL}, )
  DBUILD_REPO_URL="https://github.com/openebs/dynaic-localpv-provisioner"
  export DBUILD_REPO_URL
endif

# Specify the docker arg for website url
ifeq (${DBUILD_SITE_URL}, )
  DBUILD_SITE_URL="https://openebs.io"
  export DBUILD_SITE_URL
endif

export DBUILD_ARGS=--build-arg DBUILD_DATE=${DBUILD_DATE} --build-arg DBUILD_REPO_URL=${DBUILD_REPO_URL} --build-arg DBUILD_SITE_URL=${DBUILD_SITE_URL} --build-arg ARCH=${ARCH}

# Specify the name of base image for ARM64
ifeq (${BASE_DOCKER_IMAGE_ARM64}, )
  BASE_DOCKER_IMAGE_ARM64 = "arm64v8/ubuntu:18.04"
  export BASE_DOCKER_IMAGE_ARM64
endif

# Specify the name of base image for PPC64LE
ifeq (${BASE_DOCKER_IMAGE_PPC64LE}, )
  BASE_DOCKER_IMAGE_PPC64LE = "ubuntu:18.04"
  export BASE_DOCKER_IMAGE_PPC64LE
endif

include buildscripts/common.mk
include ./buildscripts/provisioner-localpv/Makefile.mk

.PHONY: all
all: test provisioner-localpv-image

.PHONY: all.arm64
all.arm64: provisioner-localpv-image.arm64

.PHONY: all.ppc64le
all.ppc64le: provisioner-localpv-image.ppc64le

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

.PHONY: vendor
vendor: go.mod go.sum deps
	@go mod vendor

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

.PHONY: format
format:
	@echo "--> Running go fmt"
	@go fmt $(PACKAGES) $(PACKAGES_IT)

# -composite: avoid "literal copies lock value from fakePtr"
.PHONY: vet
vet:
	@echo "--> Running go vet"
	@go list ./... | grep -v "./vendor/*" | xargs go vet -composites

.PHONY: verify-src
verify-src: 
	@echo "--> Checking for git changes post running tests";
	$(PWD)/buildscripts/check-diff.sh "format"
