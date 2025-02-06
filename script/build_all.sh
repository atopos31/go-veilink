#!/bin/bash

PLATFORMS=("linux/amd64" "linux/arm64" "windows/amd64")
OUTPUT_DIR="./bin"
APP_NAME="veilink"

# 通过 git 获取版本号
VERSION=$(git describe --tags --always)

mkdir -p "$OUTPUT_DIR"

# 遍历平台进行编译
for PLATFORM in "${PLATFORMS[@]}"
do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_CLIENT_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}-client"
    OUTPUT_SERVER_NAME="${APP_NAME}-${VERSION}-${GOOS}-${GOARCH}-server"
  
    # 为 Windows 添加 .exe 后缀
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_CLIENT_NAME+=".exe"
        OUTPUT_SERVER_NAME+=".exe"
    fi
  
    echo "building... $GOOS/$GOARCH"
    GOOS=$GOOS GOARCH=$GOARCH go build -o "${OUTPUT_DIR}/${OUTPUT_CLIENT_NAME}" ./cmd/client/client.go
    GOOS=$GOOS GOARCH=$GOARCH go build -o "${OUTPUT_DIR}/${OUTPUT_SERVER_NAME}" ./cmd/server/server.go
  
    # 检查编译结果
    if [ $? -ne 0 ]; then
        echo "build error $GOOS/$GOARCH"
    else
        echo "build success $OUTPUT_PATH"
    fi
done