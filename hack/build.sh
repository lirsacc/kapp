#!/bin/bash

set -e -x -u

go fmt ./cmd/... ./pkg/... ./test/...

build_values_path="../../../${BUILD_VALUES:-./hack/build-values-default.yml}"

(
	# template all playground assets
	# into a single Go file
	cd pkg/kapp/website; 

	ytt version || { echo >&2 "ytt is required for building. Install from https://github.com/k14s/ytt"; exit 1; }
	ytt template \
		-f . \
		-f $build_values_path \
		--file-mark 'generated.go.txt:exclusive-for-output=true' \
		--output-directory ../../../tmp/
)
mv tmp/generated.go.txt pkg/kapp/website/generated.go

go build -o kapp ./cmd/kapp/...
./kapp version

# build aws lambda binary
export GOOS=linux GOARCH=amd64
go build -o ./tmp/main ./cmd/kapp-lambda-website/...
(
	cd tmp
	chmod +x main
	rm -f kapp-lambda-website.zip
	zip kapp-lambda-website.zip main
)

echo "Success"
