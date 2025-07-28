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

# Go程序（使用cgo）
build: dict.go parse.go dict_service.h
	CGO_ENABLED=1 $(GO) build -o build/dict .

# 清理
clean:
	rm -f $(TARGETS)

.PHONY: all clean test-go test-original test-example install-deps 