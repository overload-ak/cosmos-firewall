FROM golang:alpine AS builder

LABEL stage=gobuilder

RUN apk add --no-cache git build-base linux-headers

WORKDIR /build

ADD go.mod .
ADD go.sum .
RUN go env -w GO111MODULE=on && go mod download
COPY . .
RUN go build -ldflags="-s -w" -o /app/main cmd/main.go


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/main /app/main

CMD ["./main"]
