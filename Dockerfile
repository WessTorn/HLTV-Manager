FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -buildvcs=false -o /out/HLTV-Manager .

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/HLTV-Manager /app/HLTV-Manager
COPY config.env /app/config.env
COPY hltv-runners.yaml /app/hltv-runners.yaml

EXPOSE 3000

CMD ["/app/HLTV-Manager"]
