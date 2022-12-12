FROM golang AS builder

WORKDIR /src
COPY . .
ENV CGO_ENABLED=0
RUN go build -v .

FROM alpine

EXPOSE 8080/tcp

COPY --from=builder /src/viva /bin/viva

ENTRYPOINT ["/bin/viva", "--prometheus-listen=:8080", "--viva=malm,flinten"]
