#!/bin/bash -euf -o pipefail
gitbook build
rm -rf ../docs
mv _book ../docs
