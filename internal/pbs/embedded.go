package pbs

import (
	_ "embed"

	"github.com/bufbuild/protocompile"
	"github.com/cockroachdb/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

//go:embed googleapis.pb
var googleapis []byte

//go:embed protobuf.pb
var protobuf []byte

type ThirdPartyResolver struct {
	items []*descriptorpb.FileDescriptorSet
}

func NewResolver() (*ThirdPartyResolver, error) {
	resolver := &ThirdPartyResolver{
		items: make([]*descriptorpb.FileDescriptorSet, 0, 2), //nolint:mnd
	}

	for _, pb := range [][]byte{googleapis, protobuf} {
		fds := &descriptorpb.FileDescriptorSet{}

		err := proto.Unmarshal(pb, fds)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to unmarshal descriptor: %s", pb)
		}

		resolver.items = append(resolver.items, fds)
	}

	return resolver, nil
}

func (p *ThirdPartyResolver) FindFileByPath(path string) (protocompile.SearchResult, error) {
	for _, pb := range p.items {
		for _, file := range pb.GetFile() {
			if file.GetName() == path {
				return protocompile.SearchResult{Proto: file}, nil
			}
		}
	}

	return protocompile.SearchResult{}, protoregistry.NotFound
}
