
TAG=$(shell git rev-parse HEAD)

docker:
	docker build -t thefuturezebras/project:$(TAG) -t thefuturezebras/project:dev  .
push: docker
	docker push thefuturezebras/project:$(TAG)
