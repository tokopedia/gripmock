FROM golang:1.15.3
RUN mkdir /proto

RUN mkdir /stubs

# System setup
RUN apt-get update && apt-get install -y unzip git curl

# Install protoc
ENV PATH $PATH:$(go env GOPATH)/bin
ENV PROTOBUF_URL https://github.com/protocolbuffers/protobuf/releases/download/v3.12.2/protoc-3.12.2-linux-x86_64.zip
RUN curl -L -o /tmp/protobuf.tar.gz $PROTOBUF_URL
WORKDIR /tmp/
RUN unzip protobuf.tar.gz
RUN mv /tmp/bin/protoc /usr/bin

RUN git clone --depth 1 --branch v1.3.5 https://github.com/golang/protobuf
RUN cd /tmp/protobuf/protoc-gen-go && go build -o protoc-gen-go
RUN cp /tmp/protobuf/protoc-gen-go/protoc-gen-go $(go env GOPATH)/bin
RUN rm -rf /tmp/protobuf/

RUN go get github.com/markbates/pkger/cmd/pkger

# cloning well-known-types
RUN git clone https://github.com/google/protobuf.git /protobuf-repo

RUN mkdir protobuf

# only use needed files
RUN mv /protobuf-repo/src/ /protobuf/

RUN rm -rf /protobuf-repo

RUN mkdir -p /go/src/github.com/tokopedia/gripmock

COPY . /go/src/github.com/tokopedia/gripmock

WORKDIR /go/src/github.com/tokopedia/gripmock/protoc-gen-gripmock

RUN pkger

# install generator plugin
RUN go install -v

WORKDIR /go/src/github.com/tokopedia/gripmock

# install gripmock
RUN go install -v

EXPOSE 4770 4771

ENTRYPOINT ["gripmock"]
