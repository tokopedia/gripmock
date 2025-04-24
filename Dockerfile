FROM golang:1.21-alpine

# install tools (bash, git, protobuf, protoc-gen-go, protoc-grn-go-grpc, pkger)
RUN apk -U --no-cache add bash git protobuf &&\
    go install -v github.com/golang/protobuf/protoc-gen-go@latest &&\
    go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &&\
    go install github.com/markbates/pkger/cmd/pkger@latest

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
    ln -s /go/src/github.com/tokopedia/gripmock/scripts/start_server.sh /bin/

# Copy server.go and go.mod to /go/src/grpc
RUN mkdir -p /go/src/grpc &&\
    cp /go/src/github.com/tokopedia/gripmock/scripts/server.go /go/src/grpc/ &&\
    cp /go/src/github.com/tokopedia/gripmock/scripts/go.mod /go/src/grpc/

# install plugin protoc-gen-go-grpc
WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

# install generator plugin
RUN rm -f pkged.go && pkger && go install -v

WORKDIR /go/src/github.com/tokopedia/gripmock

# Set Go version for protobuf generation
ENV GO_VERSION=1.21

# install gripmock
RUN go install -v

# Setup examples in protogen/example
RUN ./scripts/setup_examples.sh

# run server for caching purposes
RUN start_server.sh

EXPOSE 4770 4771

ENTRYPOINT ["gripmock"]
CMD ["--help"]
