FROM golang:1.17.6 as base

WORKDIR latency_checker/

COPY go.mod go.sum ./
COPY chat/ chat/
COPY cmd/ cmd/

RUN CGO_ENABLED=0 GO111MODULE=on go build -o app cmd/main.go

from alpine:3.15 as test

WORKDIR /root/

COPY --from=base go/latency_checker/app  ./

ENTRYPOINT ["./app"]
