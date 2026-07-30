FROM golang:1.26.5-alpine3.24 AS builder
WORKDIR /build
COPY . .
RUN go build -o librebread

FROM alpine:3.24
WORKDIR /app
COPY --from=builder /build/librebread .
COPY static/js/librepaymets.js /app/static/js/librepaymets.js
EXPOSE 443 80
CMD ["./librebread"]
