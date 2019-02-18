FROM golang:1.11

EXPOSE 15657

COPY ./ ./

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
RUN dep init && dep ensure

CMD [ "go", "run", "cmd/main.go"]
