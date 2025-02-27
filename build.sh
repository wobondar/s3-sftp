#!/usr/bin/env bash

bin_name="s3-sftp"

if [ -z "$1" ] || [[ ! "$1" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Error: version argument is required and must match pattern v0.0.0"
    echo
    echo "Usage: $0 <version>"
    echo
    echo "Example: $0 v0.1.0"
    exit 1
fi

platforms=("linux/amd64" "darwin/amd64" "linux/386" "linux/arm" "linux/arm64" "darwin/arm64")
mkdir -p ./build

rm -r ./build/*

for platform in "${platforms[@]}"
do
    echo "Compiling for $platform"
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=$bin_name'_'$1'_'$GOOS'_'$GOARCH
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -o ./build/$output_name cmd/$bin_name/*.go
    if [ $? -ne 0 ]; then
        echo 'An error has occurred! Aborting the script execution...'
        exit 1
    fi
done

cd build

for platform in "${platforms[@]}"
do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    input_name=$bin_name'_'$1'_'$GOOS'_'$GOARCH
    echo "Archiving binary $platform into: $input_name.zip"
    output_name=$bin_name
    mv ./$input_name ./$output_name
    zip ./$input_name.zip ./$output_name
    mv ./$output_name ./$input_name
done

echo "Generating SHA256SUMS"
shasum -a 256 *.zip > $bin_name'_'$1'_SHA256SUMS'
echo "Verifying SHA256SUMS"
shasum -a 256 -c $bin_name'_'$1'_SHA256SUMS'
cd ..