FROM golang:1.21.0-alpine3.18 as builder
WORKDIR /build
COPY . .
RUN CGO_ENABLED=0 go build -o librebread

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /build/librebread .
COPY static/js/librepaymets.js /app/static/js/librepaymets.js
EXPOSE 443 80
CMD ["./librebread"]
