FROM alpine
ARG TARGETARCH
EXPOSE 8080/tcp
COPY viva-linux-${TARGETARCH} /bin/viva
ENTRYPOINT ["/bin/viva", "--prometheus-listen=:8080"]
