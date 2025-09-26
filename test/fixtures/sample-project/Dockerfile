FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o app ./cmd/app

FROM alpine:3.19
COPY --from=build /src/app /usr/local/bin/app
ENTRYPOINT ["/usr/local/bin/app"]
