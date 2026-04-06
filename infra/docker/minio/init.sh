#!/bin/sh
set -eu

until mc alias set local http://minio:9000 "${MINIO_ROOT_USER}" "${MINIO_ROOT_PASSWORD}" >/dev/null 2>&1; do
  sleep 1
done

until mc ls local >/dev/null 2>&1; do
  sleep 1
done

for bucket in "${OBJECT_STORAGE_BUCKET_RAW}" "${OBJECT_STORAGE_BUCKET_PROCESSED}" "${OBJECT_STORAGE_BUCKET_THUMBNAILS}"; do
  mc mb --ignore-existing "local/${bucket}"
done
