FROM golang as builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY main.go ./

RUN ls -lah
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -mod=readonly -v -o server

FROM alpine
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/server /server

CMD ["/server"]
