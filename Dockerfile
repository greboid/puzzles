FROM golang:latest as builder

WORKDIR /app
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' -asmflags -trimpath -ldflags=-buildid= -o main .

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250326.0

COPY --from=builder --chown=65532 /app/main /puzzles-site
COPY ./wordlists/. /app/wordlists
EXPOSE 8080
CMD ["/puzzles-site"]
