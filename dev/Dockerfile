FROM golang:1.22-alpine
EXPOSE 8080/tcp 8081/tcp
ENV DB_URL=postgres://postgres:postgres@postgres/postgres
WORKDIR /usr/app
COPY . /usr/app
RUN go build -o pikomirror
CMD ["/usr/app/pikomirror"]