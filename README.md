# AI-Powered Binance Futures Monitor on Cloudflare Workers

这是一个基于 Go (WebAssembly) 和 Cloudflare Workers 构建的、由大型语言模型驱动的币安期货市场智能监控工具。它整合了多种市场指标，在发现潜在交易信号后，会调用外部 AI 模型进行深度分析，并最终通过飞书 (Lark) 机器人发送富有洞察力的结构化通知。

## ✨ 功能特性

- **多维策略驱动**: 内置三大信号检测器，实时监控成交量、持仓量和多空比的异常波动。
- **AI 智能分析**: 在发现信号后，调用外部 OpenAI 兼容的 API (如自定义的 Gemini 服务) 进行深度分析，生成“核心信号 + 市场背景 + 潜在影响”的结构化报告。
- **高性能核心**: 核心策略与计算逻辑由 Go 编写并编译为 WebAssembly，确保在 Cloudflare Workers 环境中的高性能和低延迟。
- **防重复通知**: 集成 Cloudflare KV 存储作为缓存，避免在1小时内重复推送相同的信号，防止信息过载。
- **定时触发与轻松部署**: 利用 Cloudflare Cron Triggers 实现分钟级定时监控，并通过 Wrangler CLI 一键部署。
- **可定制的飞书通知**: 将交易信号和 AI 分析结果格式化为精美的飞书卡片消息，直观易读。

## 🚀 技术栈

- **核心逻辑**: Go 1.19+ (编译为 WebAssembly)
- **运行环境**: Cloudflare Workers
- **缓存**: Cloudflare KV
- **AI 服务**: 任何兼容 OpenAI API 的语言模型服务
- **部署工具**: Wrangler CLI
- **通知渠道**: 飞书 (Lark)

## 🛠️ 部署与使用指南

### 1. 准备环境

- [Go](https://golang.org/dl/) (版本 >= 1.19)
- [Node.js](https://nodejs.org/) (LTS 版本)
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/install-and-update/)

### 2. 克隆与配置

```bash
git clone <你的GitHub仓库URL>
cd <项目目录>
```

### 3. 配置 Cloudflare KV 缓存

为了防止信号重复发送，你需要创建一个 KV 命名空间。

1.  **创建 KV 命名空间**:
    在你的终端运行以下命令，创建一个名为 `SIGNAL_CACHE` 的 KV 命名空间：
    ```bash
    wrangler kv:namespace create "SIGNAL_CACHE"
    ```
2.  **更新 `wrangler.toml`**:
    执行完上述命令后，Wrangler 会输出一段配置代码，类似这样：
    ```toml
    [[kv_namespaces]]
    binding = "SIGNAL_CACHE"
    id = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
    ```
    请将这段代码**完整地**粘贴到你的 `wrangler.toml` 文件中，替换掉原有的占位符部分。

### 4. 配置环境变量与密钥

本项目依赖于一些外部服务的密钥和配置。**强烈建议**使用 Wrangler 的 `secret` 命令来配置敏感信息，而不是直接写在 `wrangler.toml` 中。

1.  **设置 AI 服务密钥** (机密信息):
    ```bash
    wrangler secret put API_KEY
    ```
    然后粘贴你的 AI 服务 API Key 并回车。

2.  **设置飞书机器人 Webhook** (机密信息):
    ```bash
    wrangler secret put LARK_WEBHOOK_URL
    ```
    然后粘贴你的飞书机器人 Webhook 地址并回车。

3.  **在 `wrangler.toml` 中配置其他变量**:
    打开 `wrangler.toml` 文件，在 `[vars]` 部分填入以下非机密信息：

| 变量名                       | 描述                                     | 示例值                          |
| ---------------------------- | ---------------------------------------- | ----------------------------------- |
| `SYMBOLS`                    | 要监控的交易对，用逗号分隔               | `"BTCUSDT,ETHUSDT"`                 |
| `OPENAI_COMPATIBLE_ENDPOINT` | 你的 OpenAI 兼容 AI 服务的 URL 地址        | `"https://api.example.com/v1/chat/completions"` |
| `AI_MODEL_NAME`              | 你的 AI 服务所使用的模型名称             | `"gemini-pro"`                      |

### 5. 编译 Go 到 Wasm

在部署之前，你需要将 Go 源代码编译成 WebAssembly 模块。

```bash
# 在 Linux, macOS, 或 Git Bash on Windows
chmod +x build.sh
./build.sh

# 在 Windows PowerShell 中
# 如果遇到执行策略问题, 先运行: Set-ExecutionPolicy -Scope Process RemoteSigned
.\build.sh
```

### 6. 部署到 Cloudflare

完成以上所有配置后，一键部署你的 Worker：

```bash
wrangler deploy
```

部署完成后，Cloudflare Cron 触发器将根据 `wrangler.toml` 的设置，每15分钟自动运行一次你的智能监控机器人。
