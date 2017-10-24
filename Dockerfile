FROM golang:1.8
LABEL author="Xavier Coulon"

# MKDIR /go/src/github.com/bytesparadise/go-url-shortener
ADD . /go/src/github.com/xcoulon/go-url-shortener
RUN go install github.com/xcoulon/go-url-shortener

EXPOSE 8080

ENTRYPOINT [ "/go/bin/go-url-shortener" ]