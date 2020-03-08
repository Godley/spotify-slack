FROM golang:1.13-alpine AS build
WORKDIR /spotify-slack
COPY . /spotify-slack
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -mod vendor -o build/spotify-slack ./spotify-slack


FROM scratch
USER nobody
COPY --chown=0:0 --from=build /support/build/spotify-slack /go/bin/spotify-slack
ENTRYPOINT ["/go/bin/spotify-slack"]
