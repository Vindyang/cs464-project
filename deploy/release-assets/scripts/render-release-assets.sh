#!/usr/bin/env bash

set -euo pipefail

release_tag="${1:?usage: render-release-assets.sh <release-tag> [output-dir]}"
output_dir="${2:-deploy/release-assets/dist}"
template_dir="deploy/release-assets/templates"

mkdir -p "$output_dir"

sed "s|__OMNISHARD_TAG__|$release_tag|g" \
  "$template_dir/docker-compose.full-microservices.yml.tpl" \
  > "$output_dir/docker-compose.full-microservices.yml"

sed "s|__OMNISHARD_TAG__|$release_tag|g" \
  "$template_dir/docker-compose.single-image-microservices.yml.tpl" \
  > "$output_dir/docker-compose.single-image-microservices.yml"

echo "Rendered release assets to $output_dir"