FROM golang:1.22 as builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

COPY server /app

WORKDIR /app

RUN go mod download

RUN go install github.com/mitranim/gow@latest
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -v -o server


FROM alpine
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /app/server

USER nobody

CMD ["/app/server"]
