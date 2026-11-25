FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags '-s -w' -o /gotypes . && mkdir /data

FROM scratch

COPY --from=builder /gotypes /gotypes
# host key and players' saved data; owned by the unprivileged user below
COPY --from=builder --chown=65532:65532 /data /data
USER 65532:65532
VOLUME /data
EXPOSE 2222

# play here:     docker run -it --rm -e TERM=xterm-256color <image>
# serve on ssh:  docker run -d -p 2222:2222 -v gotypes:/data <image> -ssh :2222 -hostkey /data/hostkey -data /data
ENTRYPOINT ["/gotypes"]
