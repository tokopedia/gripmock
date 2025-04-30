FROM golang:1.23-alpine

# install tools (bash, git, protobuf, protoc-gen-go, protoc-grn-go-grpc)
RUN apk -U --no-cache add bash git protobuf &&\
    go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest &&\
    go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# cloning well-known-types
# only use needed files
RUN git clone --depth=1 https://github.com/google/protobuf.git /protobuf-repo &&\
    mv /protobuf-repo/src/ /protobuf/ &&\
    rm -rf /protobuf-repo

COPY . /go/src/github.com/tokopedia/gripmock

# create necessary dirs and export scripts
RUN mkdir -p /proto /stubs /protogen &&\
    chmod +x /go/src/github.com/tokopedia/gripmock/scripts/*.sh &&\
    ln -s /go/src/github.com/tokopedia/gripmock/scripts/fix_gopackage.sh /bin/ &&\
    ln -s /go/src/github.com/tokopedia/gripmock/scripts/start_server.sh /bin/ &&\
    ln -s /go/src/github.com/tokopedia/gripmock/scripts/wait_for_gripmock.sh /bin/

# Copy server.go and go.mod to /go/src/grpc
# to build go module
RUN mkdir -p /go/src/grpc &&\
    cp /go/src/github.com/tokopedia/gripmock/scripts/server/server.go /go/src/grpc/ &&\
    cp /go/src/github.com/tokopedia/gripmock/scripts/server/go.mod /go/src/grpc/

# install plugin protoc-gen-go-grpc
WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

# install generator plugin
RUN go install -v

WORKDIR /go/src/github.com/tokopedia/gripmock

# install gripmock
RUN go install -v

# cache the dependencies then clean up
RUN ./scripts/setup_examples.sh && \
    go build example/simple/client/*.go && \
    find . -name "*.pb.go" -type f -delete

# run server for caching purposes
RUN start_server.sh

EXPOSE 4770 4771

VOLUME /proto /stubs

ENTRYPOINT ["gripmock"]
