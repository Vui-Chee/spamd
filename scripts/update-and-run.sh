#!/bin/sh

cp frontend/index.html service/build
cp frontend/styles.css service/build

go run mdpreview.go $1