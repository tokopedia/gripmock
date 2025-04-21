#!/bin/bash

protos=("$@")

# Set sed -i parameter based on OS
if [[ "$OSTYPE" == "darwin"* ]]; then
  sed_i=(-i "")
else
  sed_i=(-i)
fi

for proto in "${protos[@]}"
do
  # if it's a directory then skip
  if [[ -d $proto ]]; then
    continue
  fi

  # example $proto: example/foo/bar/hello.proto

  # get string from left until the last /
  # example value: example/foo/bar/
  dir=${proto%/*}

  # remove prefix / if any
  dir=$(echo $dir | sed -n 's:^/*\(.*\)$:\1:p')

  # get string from right until the first /
  # example value: hello.proto
  file=${proto##*/}

  newdir="protogen/$dir"
  newfile="$newdir/$file"

  # copy to protogen directory
  mkdir -p "$newdir" && \
    cp "$proto" "$_" && \

  # Force remove any declaration of go_package
  # then replace it with our own declaration below
  sed "${sed_i[@]}" 's/^option go_package.*$//g' $newfile

  goPackageString="option go_package = \"github.com/tokopedia/gripmock/protogen/$dir\";"

  # append our own go_package delcaration just below "syntax" declaration
  sed "${sed_i[@]}" "/^syntax.*$/a $goPackageString" $newfile
  echo $newfile
done
