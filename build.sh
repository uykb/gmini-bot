
#!/bin/bash

# 此脚本用于将Go语言编写的Cloudflare Worker编译为Wasm格式

# 设置Go编译环境为Wasm
export GOOS=js
export GOARCH=wasm

# 编译Go代码
# -o main.wasm: 指定输出文件名为 main.wasm
# . (点): 表示编译当前目录下的Go代码
go build -o main.wasm .

# 检查编译是否成功
if [ $? -eq 0 ]; then
  echo "✅ Go Wasm 编译成功 -> main.wasm"
else
  echo "❌ Go Wasm 编译失败"
  exit 1
fi
