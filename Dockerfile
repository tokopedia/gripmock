FROM golang:alpine3.17

# install tools (bash, git, protobuf, protoc-gen-go, protoc-grn-go-grpc, pkger)
RUN apk -U --no-cache add bash git protobuf &&\
    go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest &&\
    go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &&\
    go install github.com/markbates/pkger/cmd/pkger@latest

# cloning well-known-types
# only use needed files
RUN git clone --depth=1 https://github.com/google/protobuf.git /protobuf-repo &&\
    mv /protobuf-repo/src/ /protobuf/ &&\
    rm -rf /protobuf-repo
# buf/validate
RUN git clone --depth=1 https://github.com/bufbuild/protovalidate.git /protobuf-repo &&\
    mv /protobuf-repo/proto/protovalidate/buf /protobuf/ &&\
    rm -rf /protobuf-repo

COPY . /go/src/github.com/tokopedia/gripmock

# create necessary dirs and export fix_gopackage.sh
RUN mkdir /proto /stubs &&\
    ln -s /go/src/github.com/tokopedia/gripmock/fix_gopackage.sh /bin/

WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

# install generator plugin
RUN pkger && go install -v

WORKDIR /go/src/github.com/tokopedia/gripmock

# install gripmock & build example to cache necessary imports
RUN go install -v && go build ./example/simple/client

EXPOSE 4770 4771

ENTRYPOINT ["gripmock"]
