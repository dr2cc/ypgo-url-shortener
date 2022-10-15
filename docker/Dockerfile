FROM golang:1.18

WORKDIR /usr/src/app

RUN apt-get -q update && apt-get -qy install netcat

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN cat tools.go | grep _ | awk -F'"' '{print $2}' | xargs -tI % go install %

RUN go build -v -o /usr/local/bin/app -ldflags "-X main.buildVersion=$(git describe --tags) -X main.buildCommit=$(git rev-parse --verify HEAD) -X 'main.buildDate=$(date +'%Y/%m/%d')'" ./cmd/shortener