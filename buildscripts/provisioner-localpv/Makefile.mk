
# Specify the name for the binaries
PROVISIONER_LOCALPV=provisioner-localpv

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
	@cd buildscripts/provisioner-localpv && sudo docker build -t openebs/provisioner-localpv:${IMAGE_TAG} --build-arg BUILD_DATE=${BUILD_DATE} . --no-cache
	@rm buildscripts/provisioner-localpv/${PROVISIONER_LOCALPV}

.PHONY: provisioner-localpv-image.arm64
provisioner-localpv-image.arm64: provisioner-localpv
	@echo "-------------------------------"
	@echo "--> provisioner-localpv image "
	@echo "-------------------------------"
	@cp bin/provisioner-localpv/${PROVISIONER_LOCALPV} buildscripts/provisioner-localpv/
	@cd buildscripts/provisioner-localpv && sudo docker build -t openebs/provisioner-localpv-arm64:${IMAGE_TAG} -f Dockerfile.arm64 --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${BASE_DOCKER_IMAGE_ARM64} . --no-cache
	@rm buildscripts/provisioner-localpv/${PROVISIONER_LOCALPV}

.PHONY: provisioner-localpv-image.ppc64le
provisioner-localpv-image.ppc64le: provisioner-localpv
	@echo "-------------------------------"
	@echo "--> provisioner-localpv image "
	@echo "-------------------------------"
	@cp bin/provisioner-localpv/${PROVISIONER_LOCALPV} buildscripts/provisioner-localpv/
	@cd buildscripts/provisioner-localpv && sudo docker build -t openebs/provisioner-localpv-ppc64le:${IMAGE_TAG} -f Dockerfile.ppc64le --build-arg BUILD_DATE=${BUILD_DATE} --build-arg BASE_IMAGE=${BASE_DOCKER_IMAGE_PPC64LE} . --no-cache
	@rm buildscripts/provisioner-localpv/${PROVISIONER_LOCALPV}
