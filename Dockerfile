FROM golang:1.22.0 as builder
ARG CGO_ENABLED=0
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build

FROM scratch
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
