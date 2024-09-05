FROM golang:1.21.0-alpine3.18 AS builder
WORKDIR /build
COPY . .
RUN go build -o librebread

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /build/librebread .
COPY static/js/librepaymets.js /app/static/js/librepaymets.js
EXPOSE 443 80
CMD ["./librebread"]
