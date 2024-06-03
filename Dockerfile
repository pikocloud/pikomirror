FROM scratch
# 8080 for Web UI, 8081 for Mirror
EXPOSE 8080/tcp 8081/tcp
ENTRYPOINT ["/pikomirror"]
ADD pikomirror /