FROM golang:latest
 COPY ./bin/* /app
 COPY ./token.txt /token.txt
 ENTRYPOINT ["/app"]
