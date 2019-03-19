
TAG=$(shell git rev-parse HEAD)
PROJECT=thefuturezebras/project

docker:
	docker build -t $(PROJECT):$(TAG) -t thefuturezebras/project:dev  .
push_docker: docker
	docker push $(PROJECT):$(TAG)

push_docker_latest: docker
	docker push $(PROJECT):latest
push_docker_dev: docker
	docker push $(PROJECT):dev

git:
	git push

commit: git push_docker


.PHONY: commit