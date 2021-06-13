FROM golang:alpine AS builder

WORKDIR /go/src/app

COPY . .

RUN ls
RUN ls data

RUN go mod tidy

RUN go get -u -d -v ./...

RUN go build -a -v -ldflags '-s -w' -o main main.go

FROM scratch

COPY --from=builder /go/src/app/ go/bin/

WORKDIR /go/bin

ENTRYPOINT ["/go/bin/main"]
