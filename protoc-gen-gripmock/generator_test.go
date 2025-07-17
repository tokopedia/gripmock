package main

import (
	"testing"

	desc "google.golang.org/protobuf/types/descriptorpb"
)

func Test_getMessageType(t *testing.T) {
	protos := []*desc.FileDescriptorProto{
		{
			Name:    strPtr("user.proto"),
			Package: strPtr("example"),
			MessageType: []*desc.DescriptorProto{
				{
					Name: strPtr("User"),
					NestedType: []*desc.DescriptorProto{
						{
							Name: strPtr("Profile"),
							NestedType: []*desc.DescriptorProto{
								{
									Name: strPtr("Address"),
								},
							},
						},
					},
				},
				{
					Name: strPtr("Profile"), // top-level Profile
				},
			},
		},
		{
			Name:    strPtr("order.proto"),
			Package: strPtr("shop"),
			MessageType: []*desc.DescriptorProto{
				{
					Name: strPtr("Order"),
				},
			},
		},
		{
			Name:    strPtr("aliased.proto"),
			Package: strPtr("aliased"),
			Options: &desc.FileOptions{
				GoPackage: strPtr("github.com/example/aliased;aliasedpkg"),
			},
			MessageType: []*desc.DescriptorProto{
				{
					Name: strPtr("AliasedMsg"),
				},
			},
		},
	}

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "top-level message",
			input: ".example.User",
			want:  "User",
		},
		{
			name:  "nested message",
			input: ".example.User.Profile",
			want:  "User_Profile",
		},
		{
			name:  "multi-level nested message",
			input: ".example.User.Profile.Address",
			want:  "User_Profile_Address",
		},
		{
			name:  "top-level Profile (ambiguous name)",
			input: ".example.Profile",
			want:  "Profile",
		},
		{
			name:  "message from another package",
			input: ".shop.Order",
			want:  "Order",
		},
		{
			name:  "aliased package message",
			input: ".aliased.AliasedMsg",
			want:  "aliasedpkg.AliasedMsg",
		},
		{
			name:  "unknown message fallback",
			input: ".example.Unknown",
			want:  "Unknown",
		},
		{
			name:  "empty input",
			input: "",
			want:  "",
		},
		{
			name:  "dot only input",
			input: ".",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getMessageType(protos, tt.input)
			if got != tt.want {
				t.Errorf("getMessageType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func strPtr(s string) *string { return &s }
