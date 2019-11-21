FROM golang:buster as builder

WORKDIR /go-short
COPY . .
RUN CGO_ENABLED=0 go build -a -o go-short

FROM alpine:latest

RUN mkdir $HOME/.go-short
COPY resources/go-short.yml /root/.go-short

WORKDIR /go-short

COPY --from=builder /go-short/go-short .

EXPOSE 8080

CMD ["./go-short" , "server"]