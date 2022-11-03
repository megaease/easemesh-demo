.PHONY: build build_docker

build:
	mvn clean package -Dmaven.test.skip

build_docker: build
	docker buildx build --platform linux/amd64 --load -t megaease/easemesh-demo-restaurant:latest .
