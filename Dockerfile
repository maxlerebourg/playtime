FROM golang:1.26-alpine AS go

RUN mkdir build

WORKDIR build
COPY . .
RUN CGO_ENABLED=0 go build -o /playtime .


FROM alpine

COPY --from=go /playtime /playtime

RUN mkdir -m 0777 /data /uploads

ENTRYPOINT ["/playtime"]
