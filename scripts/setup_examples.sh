#!/bin/bash

# Create protogen/example directory
rm -rf protogen/example
mkdir -p protogen/example

# Find all proto files in current directory and process them with fix_gopackage
proto_files=($(find . -name "*.proto" | sed 's|^\./||'))
echo "Found proto files: ${proto_files[@]}"

# Process all proto files with fix_gopackage
./scripts/fix_gopackage.sh "${proto_files[@]}"

# Change to protogen/example directory
cd protogen/example

# Process each example directory
for dir in */; do
    if [ -d "$dir" ]; then
        echo "Processing $dir..."
        
        # Change to example directory
        cd "$dir"
        
        # Find all .proto files recursively and process them together
        proto_files=($(find . -name "*.proto"))
        echo "proto_files: ${proto_files[@]}"
        if [ ${#proto_files[@]} -gt 0 ]; then
            echo "Generating protobuf for ${#proto_files[@]} proto files..."
            protoc --go_out=. --go_opt=paths=source_relative \
                   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
                   "${proto_files[@]}"
        fi
        
        # Go back to parent directory
        cd ..
    fi
done

# remove everything except .pb.go files
find . -type f ! -name '*.pb.go' -delete

# remove all empty directories
find . -type d -empty -delete

# Go back to root
cd ../..

echo "Setup complete!" 