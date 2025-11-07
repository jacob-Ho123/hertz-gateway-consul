# 定义变量
DOCKER_REGISTRY = jacob157
DOCKER_IMAGE_NAME = $(DOCKER_REGISTRY)/hertz-gateway
DOCKER_IMAGE_TAG = v0.0.1
DOCKER_CONTAINER_NAME = hertz-gateway
DOCKER_PORT = 80

# 构建Docker镜像
build:
	DOCKER_BUILDKIT=1 docker build -t $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG) .

build_test:
	DOCKER_BUILDKIT=1 docker build -t $(DOCKER_IMAGE_NAME):test .
# 运行Docker容器
run:
	docker run -d --name $(DOCKER_CONTAINER_NAME) -p $(DOCKER_PORT):9000 $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

# 停止并删除Docker容器
stop:
	docker stop $(DOCKER_CONTAINER_NAME)
	docker rm $(DOCKER_CONTAINER_NAME)

# 删除Docker镜像
clean:
	docker rmi $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

# 重新构建并运行
rebuild: stop clean build run

# 查看容器日志
logs:
	docker logs $(DOCKER_CONTAINER_NAME)

# 进入容器shell
shell:
	docker exec -it $(DOCKER_CONTAINER_NAME) /bin/bash

# 推送Docker镜像到仓库
push:
	docker push $(DOCKER_IMAGE_NAME):$(DOCKER_IMAGE_TAG)

push_test:
	docker push $(DOCKER_IMAGE_NAME):test

.PHONY: build build_test push_test run stop clean rebuild logs shell push
