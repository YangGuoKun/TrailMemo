# TrailMemo 智能旅行路线与打卡平台

TrailMemo 是一个基于 Go + Gin + GORM + MySQL + Redis + UniApp/Vue3 的智能旅行路线与打卡平台。项目覆盖用户认证、路线规划、打卡记录、社区分享等常规业务，并在后端内嵌 Go-native AI Agent 模块，支持自然语言生成路线草稿、旅行推荐、路线改造、游记生成和 SSE 流式对话。



## 技术栈

**后端**

- Go / Gin / GORM
- MySQL / Redis
- JWT 鉴权
- RESTful API / SSE
- Zap 日志
- Swagger API 文档
- LLM API
- 高德地图 POI API
- JSON Schema 参数校验

**前端**

- UniApp
- Vue3
- Pinia
- TypeScript
- uView Plus
- 微信小程序

**AI Agent**

- Go-native Agent
- Workflow-first 编排
- Tool Registry
- Run / Step / Artifact 运行记录
- Artifact + Approval Gate
- Agent Eval golden cases

## 核心功能

- 用户注册、登录、JWT 鉴权
- 用户资料、头像上传
- 路线创建、路线详情、路线状态管理
- 打卡点管理与打卡记录
- 社区帖子、评论、点赞、收藏
- AI 路线草稿生成
- AI 旅行推荐
- AI 路线改造
- AI 游记生成
- Agent SSE 流式对话
- Agent 运行详情、失败步骤和 token 统计查询

## Agent 架构亮点

### 1. Workflow-first Agent

TrailMemo 没有采用完全自治 Agent，而是使用 Workflow-first 架构。路线生成、旅行推荐、路线改造、游记生成等能力都有明确的业务链路，因此通过固定工作流约束 LLM 的行为：

```text
用户输入
-> Agent 意图识别
-> Route Draft Workflow
-> LLM 提取路线约束
-> Tool Registry 查询公开路线和高德 POI
-> LLM 生成结构化 route_draft
-> 保存 Agent Artifact
-> 前端卡片确认
-> Approve + Commit
-> RouteService 创建真实 Route 和 Checkpoints
```

这样既能发挥 LLM 的自然语言理解和生成能力，又避免它绕过业务规则直接操作数据库。

### 2. Tool Registry

Agent 通过统一的 Tool Registry 调用业务工具，而不是直接访问数据库。每个工具具备明确的名称、权限等级、参数 Schema 和执行入口。

主要能力包括：

- 工具注册与统一查找
- 工具权限分级
- JSON Schema 参数校验
- 工具执行结果结构化返回
- 对路线、打卡、社区、高德 POI 等服务的受控访问

### 3. Artifact + Approval Gate

Agent 生成的路线不会立即写入用户数据，而是先保存为 `route_draft` Artifact。前端将 Artifact 渲染成路线卡片，用户点击确认后，后端再通过 Commit 流程调用 `RouteService` 创建真实路线和打卡点。

这个设计把 AI 输出和用户数据写入隔离开：

- AI 负责生成草稿
- 用户负责确认意图
- 后端服务负责最终落库

### 4. Run / Step / Artifact 可观测性

Agent 每次执行会记录运行过程：

- `agent_runs`：一次 Agent 运行记录
- `agent_steps`：意图识别、LLM 调用、工具调用、Artifact 生成等步骤
- `agent_artifacts`：结构化产物，例如路线草稿、游记草稿

这些记录支持 `run_id` 查询、失败步骤定位、token 统计、latency 统计和前端运行详情展示，便于调试 Agent 行为。

## 项目结构

```text
TrailMemo
├── trailmemo/                 # Go 后端服务
│   ├── cmd/server/            # 服务启动入口
│   ├── configs/               # 后端配置文件，本地 config.yaml 不应提交
│   ├── internal/
│   │   ├── agent/             # Go-native AI Agent 模块
│   │   │   ├── handler/       # Agent HTTP 接口
│   │   │   ├── service/       # Agent 应用服务
│   │   │   ├── workflow/      # Workflow-first 编排
│   │   │   ├── tools/         # Tool Registry 与业务工具
│   │   │   ├── memory/        # Run/Step/Artifact 存储
│   │   │   ├── llm/           # LLM 客户端
│   │   │   ├── guardrail/     # 输入输出安全校验
│   │   │   └── eval/          # Agent golden cases
│   │   ├── config/            # MySQL、Redis、配置加载
│   │   ├── handler/           # 业务 HTTP Handler
│   │   ├── middleware/        # JWT、CORS、日志中间件
│   │   ├── model/             # GORM 数据模型
│   │   ├── repository/        # 数据访问层
│   │   └── service/           # 业务服务层
│   └── docs/                  # Swagger 文档
├── TrailMemo-app/             # UniApp + Vue3 前端
│   ├── src/api/               # 前端 API 封装
│   ├── src/pages/             # 页面
│   ├── src/stores/            # Pinia 状态管理
│   └── src/utils/             # Agent 意图识别、Markdown 渲染等工具
├── doc/                       # 架构设计与学习文档
└── trailmemo-rag/             # RAG/Agent 实验目录
```

## 本地启动

### 1. 环境准备

- Go 1.25+
- MySQL 8+
- Redis 6+
- Node.js 18+
- 微信开发者工具

### 2. 后端配置

后端配置文件位于：

```text
trailmemo/configs/config.yaml
```

该文件包含数据库密码、JWT Secret、微信 AppSecret、LLM Key、高德地图 Key 等敏感信息，已被 `.gitignore` 忽略。首次运行时请在本地创建该文件，并按自己的环境填写配置。

推荐使用环境变量覆盖敏感配置：

```bash
TRAILMEMO_LLM_API_KEY=your_llm_api_key
TRAILMEMO_AGENT_LLM_API_KEY=your_agent_llm_api_key
TRAILMEMO_MAP_API_KEY=your_amap_key
TRAILMEMO_WECHAT_APPID=your_wechat_appid
TRAILMEMO_WECHAT_APPSECRET=your_wechat_appsecret
```

### 3. 启动后端

```bash
cd trailmemo
go mod download
go run ./cmd/server
```

默认服务地址：

```text
http://localhost:8087
```

健康检查：

```text
GET http://localhost:8087/health
```

Swagger 文档：

```text
http://localhost:8087/swagger/index.html
```

### 4. 启动前端

```bash
cd TrailMemo-app
npm install
npm run dev:mp-weixin
```

然后使用微信开发者工具导入生成的小程序项目进行预览和调试。

H5 调试：

```bash
npm run dev:h5
```

## 测试

后端测试：

```bash
cd trailmemo
go test ./...
```

前端工具测试：

```bash
cd TrailMemo-app
node scripts/agentIntent.test.mjs
node scripts/agentFailure.test.mjs
node scripts/markdown.test.mjs
```

微信小程序构建：

```bash
cd TrailMemo-app
npm run build:mp-weixin
```

## 主要接口

基础路径：

```text
/api/v1
```

常见模块：

- `/users`：用户注册、登录、资料
- `/routes`：路线管理
- `/checkins`：打卡记录
- `/posts`：社区帖子
- `/comments`：评论
- `/agent`：Agent 对话、Workflow、Run 查询、Artifact 审批提交

完整接口以 Swagger 文档和代码路由为准。




## 项目状态

该项目目前适合作为 Go 后端 / AI 应用开发学习项目展示，重点体现：

- 后端分层架构能力
- Gin + GORM + Redis 的业务开发能力
- LLM 接入真实业务系统的工程化能力
- Agent Workflow 编排与工具调用设计能力
- AI 写操作安全边界设计能力
- Agent 运行过程可观测性建设能力

