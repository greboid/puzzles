FROM ghcr.io/greboid/dockerfiles/golang:latest as builder

WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -asmflags -trimpath -ldflags=-buildid= -o main .

FROM ghcr.io/greboid/dockerfiles/base:latest

COPY --from=builder --chown=65532 /app/main /puzzles-site
COPY ./wordlists/. /app/wordlists
EXPOSE 8080
CMD ["/puzzles-site"]
