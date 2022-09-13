FROM golang:1.18

WORKDIR /usr/src/app

RUN apt-get -q update && apt-get -qy install netcat

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go install github.com/golang/mock/mockgen@v1.6.0
RUN go install mvdan.cc/gofumpt@latest
RUN go install golang.org/x/tools/cmd/godoc@latest

COPY . .
RUN go build -v -o /usr/local/bin/app ./cmd/shortener