#!/usr/bin/env bash

dir=$(cd -P -- "$(dirname -- "$0")" && pwd -P)
cd "${dir}"/../../ || exit

# shellcheck disable=SC2125
FILES=schemas/*
for f in ${FILES}
do
    FILENAME=$(basename -- "$f")
    SCHEMA="${FILENAME%.*}"

    http --ignore-stdin -v POST :8081/subjects/"${SCHEMA}"/versions \
      Accept:application/vnd.schemaregistry.v1+json \
      schema=@"${f}"
done