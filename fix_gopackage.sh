#!/bin/bash

protos=("$@")

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

  existingGoPackage=$(grep -n '^option go_package.*$' "$newfile")

  if [[ "$existingGoPackage" != *"// external"* ]]; then 
  
    # Force remove any declaration of go_package
    # then replace it with our own declaration below
    sed -i 's/^option go_package.*$//g' $newfile


  # get the line number of "syntax" declaration
  syntaxLineNum="$(grep -n "syntax" "$newfile" | head -n 1 | cut -d: -f1)"

  goPackageString="option go_package = \"github.com/tokopedia/gripmock/protogen/$dir\";"

  # append our own go_package delcaration just below "syntax" declaration
  sed -i "${syntaxLineNum}s~$~\n$goPackageString~" $newfile

  fi;

  echo $newfile
done

