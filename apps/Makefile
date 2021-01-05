
.PHONY: docker_build_images
docker_build_images:
	docker image build -t open-telem-demo/svc-one --build-arg SVC_PATH=svc-one .
	docker image build -t open-telem-demo/client-api --build-arg SVC_PATH=client-api .
