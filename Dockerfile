FROM golang:alpine

RUN mkdir /proto

RUN mkdir /stubs

RUN apk -U --no-cache add git protobuf bash

RUN go install -v github.com/golang/protobuf/protoc-gen-go@latest

RUN go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

RUN go install github.com/markbates/pkger/cmd/pkger@latest

# cloning well-known-types
RUN git clone --depth=1 https://github.com/google/protobuf.git /protobuf-repo

RUN mkdir protobuf

# only use needed files
RUN mv /protobuf-repo/src/ /protobuf/

RUN rm -rf /protobuf-repo

RUN mkdir -p /go/src/github.com/tokopedia/gripmock

COPY . /go/src/github.com/tokopedia/gripmock

RUN ln -s /go/src/github.com/tokopedia/gripmock/fix_gopackage.sh /bin/

WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

RUN pkger

# install generator plugin
RUN go install -v

WORKDIR /go/src/github.com/tokopedia/gripmock

# install gripmock
RUN go install -v

# to cache necessary imports
RUN go build ./example/simple/client

# remove all .pb.go generated files
# since generating go file is part of the test
RUN find . -name "*.pb.go" -delete -type f

EXPOSE 4770 4771

ENTRYPOINT ["gripmock"]
