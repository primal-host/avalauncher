FROM golang:1.25-alpine AS build
WORKDIR /src
ENV GOPROXY=http://host.docker.internal:3000|direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /avalauncher ./cmd/avalauncher

FROM alpine:3.21
RUN apk add --no-cache ca-certificates openssh-client
COPY --from=build /avalauncher /usr/local/bin/avalauncher
ENTRYPOINT ["avalauncher"]
