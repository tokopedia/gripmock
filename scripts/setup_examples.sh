#!/bin/bash

# Create protogen/example directory
mkdir -p protogen/example

# Copy all examples to protogen/example
cp -r example/* protogen/example/

# Change to protogen/example directory
cd protogen/example

# Process each example directory
for dir in */; do
    if [ -d "$dir" ]; then
        echo "Processing $dir..."
        
        # Change to example directory
        cd "$dir"
        
        # Generate .pb.go files if proto file exists
        if [ -f "${dir%/}.proto" ]; then
            protoc --go_out=. --go_opt=paths=source_relative \
                   --go-grpc_out=. --go-grpc_opt=paths=source_relative \
                   "${dir%/}.proto"
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