FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./

RUN go build -o /app/merchant-server .

FROM alpine

WORKDIR /app

COPY --from=builder /app/merchant-server ./merchant-server

CMD ["./merchant-server"]