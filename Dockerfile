FROM golang:1.11

EXPOSE 15657
WORKDIR $GOPATH/src/github.com/TTK4145-students-2019/project-thefuturezebras

ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main main.go

CMD [ "./main"]
