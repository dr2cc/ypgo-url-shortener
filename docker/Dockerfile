FROM golang:1.18

WORKDIR /usr/src/app

RUN apt-get -q update && apt-get -qy install netcat

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify
RUN go install github.com/golang/mock/mockgen@v1.6.0
RUN go install mvdan.cc/gofumpt@latest
RUN go install golang.org/x/tools/cmd/godoc@latest
RUN go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest

COPY . .
RUN go build -v -o /usr/local/bin/app -ldflags "-X main.buildVersion=$(git describe --tags) -X main.buildCommit=$(git rev-parse --verify HEAD) -X 'main.buildDate=$(date +'%Y/%m/%d')'" ./cmd/shortener