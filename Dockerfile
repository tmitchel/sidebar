FROM golang:1.13
WORKDIR /go/src/sidebar
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD [ "chat" ]

EXPOSE 8080
