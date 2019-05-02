FROM golang:1.6

EXPOSE 15657
WORKDIR $GOPATH/src/github.com/HaavardM/TTK4145-Elevator

ADD . .
RUN go get
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o main main.go

CMD [ "./main"]
