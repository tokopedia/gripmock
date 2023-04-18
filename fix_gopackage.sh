#!/bin/bash
#
# This script pre-processes the .proto files input to 'gripmock' to rewrite
# their "option go_package" directives so that they appear to be under the
# gripmock package.
#

set -e -u -x

echo 1>&2 "fix workdir=${PWD}"

protogen_dir=protogen

# We deliberately use a package path that does not really exist, so that "go
# mod tidy" etc will only be able to use real paths. Vendoring will be used
# to ensure the packages are present where expected.
protogen_package=${GENERATED_MODULE_NAME:-gripmock/generated}

protos=("$@")

if [ $# -eq 0 ]; then
    echo "Usage: $0 a.proto [b.proto [..]]"
	exit 1
fi

# All generated protocols are under a "protogen" dir, which must be a separate
# package. There must be an empty go.mod to identify it so that child packages
# are found correctly. It doesn't need to be under the gripmock source tree.
mkdir -p $protogen_dir
if ! [ -e $protogen_dir/go.mod ]; then
  (cd $protogen_dir && go mod init "$protogen_package")
fi
if ! [ -e $protogen_dir/go.sum ]; then
  touch $protogen_dir/go.sum
fi

for proto in "${protos[@]}"
do
  echo 1>&2 "proto=${proto}"
  # if it's a directory then skip
  if [[ -d $proto ]]; then
    continue
  fi

  # example $proto: example/foo/bar/hello.proto

  # Convert protocol path to a relative path within the workdir. We should in
  # future refine this prefix matching stuff.
  rel_proto="$(realpath --relative-to="." "$proto")"

  if [[ "$proto" == "*/../*" ]]; then
      echo 1>&2 "Error: protocol file \"$proto\" must be in or under directory \"$(realpath "${protogen_dir}/..")\""
      exit 1
  fi

  # Split relative-ized proto path into directory and file components
  dir="$(dirname $rel_proto)"
  file="$(basename $rel_proto)"
  echo 1>&2 "Will use package ${protogen_package}/${dir} for ${file}"

  # Outputs will be under the same relative path as the .proto file, under the
  # protogen dir.
  newdir="$protogen_dir/$dir"
  newfile="$newdir/$file"

  # copy protocol into protogen directory, transforming it to replace the
  # go_package declaration with one we generate to reflect the new directory
  # structure.
  mkdir -p "$newdir"

  # Copy the protocol file into protogen/, transforming any go_package line to
  # put the protocol package under $protogen_package to match the directory
  # structure.
  gawk \
    -v "newpackage=${protogen_package}/${dir}" \
    '
    BEGIN {
      looking_for_syntax=1;
    }
    # search for "syntax" line, which specs say must be first non-empty
    # non-comment line, and append a go_package line we generate
    # immediately after it
    looking_for_syntax && !/^syntax/ {
      print;
      next;
    }
    # Retain the "syntax" line, and append a generated go_package line
    # immediately after it.
    /^syntax/ {
      print;
      looking_for_syntax=0;
      printf("option go_package = \"%s\";\n", newpackage);
      next;
    }
    # if there was any existing go_package line, omit it, since we wrote our
    # own replacement.
    /^option[ \t]+go_package\>/ {
        next;
    }
    # Retain everything else verbatim
    { print; next; }
    ' \
    $proto > $newfile

  echo 1>&2 "newfile=${newfile}"

  echo $newfile
done


# vim: sw=2 ts=2 et ai
