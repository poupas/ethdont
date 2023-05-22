FROM golang:1.20 as builder
ADD . /build
RUN cd /build \
    && go build .

FROM debian:bullseye-slim
COPY --from=builder /build/ethdont /usr/local/bin
ENTRYPOINT ["/usr/local/bin/ethdont"]
