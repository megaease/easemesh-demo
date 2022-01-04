.PHONY: build build_docker

build:
	mvn clean package -Dmaven.test.skip

build_docker: build
	docker build -t megaease/easemesh-demo-restaurant:latest .
