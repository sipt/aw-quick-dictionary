# Makefile for Quick Dictionary Service

# 编译器设置
CC = clang
GO = go

# 编译标志
CFLAGS = -x objective-c -framework Foundation -framework CoreServices
LDFLAGS = -framework Foundation -framework CoreServices

# 目标文件
TARGETS = dict

# 默认目标
all: $(TARGETS)

# Go程序（使用cgo）build amd64 & arm64, 使用 lipo 合成一个通用二进制文件
build: dict.go parse.go dict_service.h
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO) build -o build/dict_amd64 .
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO) build -o build/dict_arm64 .
	lipo -create -output build/dict build/dict_amd64 build/dict_arm64
	rm -f build/dict_amd64 build/dict_arm64

# 清理
clean:
	rm -f $(TARGETS)

.PHONY: all clean test-go test-original test-example install-deps 