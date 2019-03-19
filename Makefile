
TAG=$(shell git rev-parse HEAD)
PROJECT=thefuturezebras/project

docker:
	docker build -t $(PROJECT):$(TAG) -t thefuturezebras/project:dev  .
push: docker
	docker push $(PROJECT):$(TAG)

push_latest: docker
	docker push $(PROJECT):latest
push_dev: docker
	docker push $(PROJECT):dev