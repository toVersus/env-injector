FROM golang:1.14.4-alpine3.12 AS builder

RUN apk --no-cache add git

RUN mkdir /app
WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ /app/cmd
COPY pkg/ /app/pkg

RUN go build -o /bin/ /app/cmd/env-injector

FROM alpine:3.12.0
RUN apk --no-cache add ca-certificates

RUN addgroup -S -g 1000 injector && adduser -S -u 1000 -G injector injector
USER injector

COPY --from=builder /bin/env-injector /usr/local/bin/env-injector

CMD ["/usr/local/bin/env-injector"]
