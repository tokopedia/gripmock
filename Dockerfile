FROM golang:alpine

RUN mkdir /proto

RUN mkdir /stubs

RUN --mount=type=cache,target=/var/cache/apk \
  apk -U add git protobuf bash gawk coreutils

RUN \
  go install -v google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
  go install -v google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# there must be a better way to fetch the protobuf well-known types than this
RUN --mount=type=cache,target=/protobuf-repo \
  [ -e /protobuf-repo/.git ] || git clone --depth=1 https://github.com/google/protobuf.git /protobuf-repo && \
  cp -r /protobuf-repo/src/ /protobuf/

ADD fix_gopackage.sh /bin/fix_gopackage.sh

RUN mkdir -p /gripmock
COPY . /gripmock

RUN ls /gripmock
RUN ls /gripmock/server_template

# Pre-download modules used for server builds
RUN cd /gripmock/server_template && \
    go mod download -x

# install generator plugin
WORKDIR /gripmock/protoc-gen-gripmock
RUN go install -v && \
    rm -rf vendor &>/dev/null

# install gripmock
WORKDIR /gripmock
RUN go install -v && \
    rm -rf vendor >&/dev/null

WORKDIR /
RUN mkdir /protogen

RUN go version

WORKDIR /

# remove all .pb.go generated files
# since generating go file is part of the test
RUN find /gripmock -name "*.pb.go" -delete -type f

EXPOSE 4770 4771

ENTRYPOINT ["gripmock", "--template-dir=/gripmock/server_template"]
