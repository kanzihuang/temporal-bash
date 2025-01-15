#!/usr/bin/env bash

project_dir=$(cd "$(dirname "$0")/.." && pwd)
# shellcheck disable=SC2016
go run "$project_dir/main.go" \
    worker -t image-copy-huaweicloud \
    -a 'skopeo-copy=skopeo copy "$source" "$destination"' \
    -a 'gzip=gzip "$file"' \
    -a 'obsutil-copy=obsutil cp "$source" "$destination"' \
    -a 'obsutil-share=obsutil create-share -ac "$access_code" -vp "$validity"'