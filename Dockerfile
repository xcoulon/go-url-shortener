# build image
FROM centos:7 as builder
LABEL author="Xavier Coulon <xcoulon@redhat.com>"
ENV LANG=en_US.utf8

# Install wget and git
RUN yum  --enablerepo=centosplus install -y \
    wget \
    git

# install golang 1.9
ENV GOLANG_VERSION=1.9.2
RUN wget -O /opt/go${GOLANG_VERSION}.linux-amd64.tar.gz https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf /opt/go${GOLANG_VERSION}.linux-amd64.tar.gz 
ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go

# install 'dep' for Go package management
RUN go get -u github.com/golang/dep/cmd/dep

# import source from host
ADD . /go/src/github.com/xcoulon/go-url-shortener
WORKDIR /go/src/github.com/xcoulon/go-url-shortener
RUN $GOPATH/bin/dep ensure -v

# run the tests, using build args to specify the connection settings to the Postgres DB
# optional args that can be filled with `build-arg` when executing the `docker build` command
ARG POSTGRES_HOST
ARG POSTGRES_PORT
ARG POSTGRES_USER
ARG POSTGRES_PASSWORD
RUN LOG_LEVEL=debug go test ./...

# build the application
ARG BUILD_COMMIT=unknown
ARG BUILD_TIME=unknown
RUN go build -ldflags "-X github.com/xcoulon/go-url-shortener/configuration.BuildCommit=${BUILD_COMMIT} -X github.com/xcoulon/go-url-shortener/configuration.BuildTime=${BUILD_TIME}" -o bin/go-url-shortener

# final image
FROM centos:7
LABEL author="Xavier Coulon <xcoulon@redhat.com>"

ARG BUILD_COMMIT=unknown
ARG BUILD_TIME=unknown
LABEL url-shortener.version=${BUILD_COMMIT} \
      url-shortener.build-time=${BUILD_TIME}

# Add the binary file generated in the `builder` container above
COPY --from=builder /go/src/github.com/xcoulon/go-url-shortener/bin/go-url-shortener /usr/local/bin/go-url-shortener

# Create a non-root user and a group with the same name: "shortenerapp"
ENV USER_GROUP=shortenerapp
RUN groupadd -r ${USER_GROUP} && \
    useradd --no-create-home -g ${USER_GROUP} ${USER_GROUP} 
# From here onwards, any RUN, CMD, or ENTRYPOINT will be run under the following user instead of 'root'
USER ${USER_GROUP} 

EXPOSE 8080

ENTRYPOINT [ "/usr/local/bin/go-url-shortener" ]