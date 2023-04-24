FROM golang:1.20 as builder

RUN groupadd -g 10001 nonroot
RUN useradd -u 10001 -g 10001 -d /app nonroot

WORKDIR /build
COPY * ./
RUN go build -o ort-operator-api .


FROM scratch

WORKDIR /app
COPY LICENSE ./
COPY --from=builder /build/ort-operator-api ./
COPY --from=builder /etc/passwd /etc/passwd

USER nonroot
EXPOSE 4000
CMD ["/app/ort-operator-api"]
