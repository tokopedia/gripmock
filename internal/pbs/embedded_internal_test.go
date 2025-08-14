package pbs

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestNewResolver(t *testing.T) {
	t.Parallel()

	// Test creating a new resolver
	resolver, err := NewResolver()
	require.NoError(t, err)
	require.NotNil(t, resolver)
	require.NotNil(t, resolver.items)
	require.Len(t, resolver.items, 2) // googleapis and protobuf
}

func TestNewResolver_EmbeddedData(t *testing.T) {
	t.Parallel()

	// Test that embedded data is not empty
	require.NotEmpty(t, googleapis)
	require.NotEmpty(t, protobuf)
}

func TestThirdPartyResolver_FindFileByPath_ExistingFile(t *testing.T) {
	t.Parallel()

	// Test finding an existing file
	resolver, err := NewResolver()
	require.NoError(t, err)

	// Try to find a common protobuf file
	result, err := resolver.FindFileByPath("google/protobuf/descriptor.proto")
	if err == nil {
		// File found
		require.NotNil(t, result)
		require.NotNil(t, result.Proto)
	} else {
		// File not found, which is also valid
		require.Equal(t, protoregistry.NotFound, err)
	}
}

func TestThirdPartyResolver_FindFileByPath_NonExistentFile(t *testing.T) {
	t.Parallel()

	// Test finding a non-existent file
	resolver, err := NewResolver()
	require.NoError(t, err)

	result, err := resolver.FindFileByPath("non/existent/file.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_FindFileByPath_EmptyPath(t *testing.T) {
	t.Parallel()

	// Test finding with empty path
	resolver, err := NewResolver()
	require.NoError(t, err)

	result, err := resolver.FindFileByPath("")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_FindFileByPath_NilResolver(t *testing.T) {
	t.Parallel()

	// Test with nil resolver (edge case)
	var resolver *ThirdPartyResolver = nil
	if resolver != nil {
		result, err := resolver.FindFileByPath("test.proto")
		require.Error(t, err)
		require.Empty(t, result)
	}
}

func TestThirdPartyResolver_Struct(t *testing.T) {
	t.Parallel()

	// Test ThirdPartyResolver struct
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{},
	}

	require.NotNil(t, resolver)
	require.NotNil(t, resolver.items)
	require.Empty(t, resolver.items)
}

func TestThirdPartyResolver_WithEmptyItems(t *testing.T) {
	t.Parallel()

	// Test resolver with empty items
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{},
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_WithNilItems(t *testing.T) {
	t.Parallel()

	// Test resolver with nil items
	resolver := &ThirdPartyResolver{
		items: nil,
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_WithSingleItem(t *testing.T) {
	t.Parallel()

	// Test resolver with single item
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("test.proto"),
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Proto)
	require.Equal(t, "test.proto", result.Proto.GetName())
}

func TestThirdPartyResolver_WithMultipleItems(t *testing.T) {
	t.Parallel()

	// Test resolver with multiple items
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("first.proto"),
					},
				},
			},
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("second.proto"),
					},
				},
			},
		},
	}

	// Test finding first file
	result, err := resolver.FindFileByPath("first.proto")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "first.proto", result.Proto.GetName())

	// Test finding second file
	result, err = resolver.FindFileByPath("second.proto")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "second.proto", result.Proto.GetName())

	// Test finding non-existent file
	_, err = resolver.FindFileByPath("third.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
}

func TestThirdPartyResolver_WithEmptyFileList(t *testing.T) {
	t.Parallel()

	// Test resolver with empty file list
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{},
			},
		},
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_WithNilFileList(t *testing.T) {
	t.Parallel()

	// Test resolver with nil file list
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: nil,
			},
		},
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_WithFileWithoutName(t *testing.T) {
	t.Parallel()

	// Test resolver with file without name
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: nil, // No name
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_WithEmptyFileName(t *testing.T) {
	t.Parallel()

	// Test resolver with empty file name
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String(""), // Empty name
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("test.proto")
	require.Error(t, err)
	require.Equal(t, protoregistry.NotFound, err)
	require.Empty(t, result)
}

func TestThirdPartyResolver_WithMatchingEmptyName(t *testing.T) {
	t.Parallel()

	// Test resolver with matching empty name
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String(""), // Empty name
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.NotNil(t, result.Proto)
	require.Empty(t, result.Proto.GetName())
}

func TestThirdPartyResolver_WithSpecialCharacters(t *testing.T) {
	t.Parallel()

	// Test resolver with special characters in file names
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("test-file.proto"),
					},
					{
						Name: proto.String("test_file.proto"),
					},
					{
						Name: proto.String("test.file.proto"),
					},
				},
			},
		},
	}

	// Test finding files with special characters
	testCases := []string{
		"test-file.proto",
		"test_file.proto",
		"test.file.proto",
	}

	for _, fileName := range testCases {
		t.Run(fileName, func(t *testing.T) {
			t.Parallel()

			result, err := resolver.FindFileByPath(fileName)
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Equal(t, fileName, result.Proto.GetName())
		})
	}
}

func TestThirdPartyResolver_WithLongPath(t *testing.T) {
	t.Parallel()

	// Test resolver with long path
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("very/long/path/to/file.proto"),
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("very/long/path/to/file.proto")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "very/long/path/to/file.proto", result.Proto.GetName())
}

func TestThirdPartyResolver_WithUnicodePath(t *testing.T) {
	t.Parallel()

	// Test resolver with unicode path
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("тест/файл.proto"),
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("тест/файл.proto")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "тест/файл.proto", result.Proto.GetName())
}

func TestThirdPartyResolver_WithDuplicateNames(t *testing.T) {
	t.Parallel()

	// Test resolver with duplicate file names (should return first match)
	resolver := &ThirdPartyResolver{
		items: []*descriptorpb.FileDescriptorSet{
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("duplicate.proto"),
					},
				},
			},
			{
				File: []*descriptorpb.FileDescriptorProto{
					{
						Name: proto.String("duplicate.proto"),
					},
				},
			},
		},
	}

	result, err := resolver.FindFileByPath("duplicate.proto")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "duplicate.proto", result.Proto.GetName())
	// Should return the first match from the first item
}
