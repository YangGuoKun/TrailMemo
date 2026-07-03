# 🤖 AI Agent 分阶段实施计划

> 基于 `AI_AGENT_DESIGN.md` 架构设计 · 面向学习与面试 · 每个 Phase 独立可演示

---

## 阶段总览

```
Phase 1 ████████████░░░░░░░░  基础框架 (2周)
Phase 2 ████████████░░░░░░░░  工具+记忆 (3周)  ← 核心亮点
Phase 3 ████████░░░░░░░░░░░░  向量+流式 (2周)  ← 面试加分
Phase 4 ████░░░░░░░░░░░░░░░░  监控+运营 (1周)
```

| 阶段 | 时间 | 面试卖点 | 可演示 |
|------|------|---------|--------|
| P1: LLM 基础设施 | 2 周 | 多服务商切换 + 提示词工程 + 降级 | ✅ 推荐接口 |
| P2: Agent 核心 | 3 周 | Tool Calling + 分层记忆 + 多步推理 | ✅ 攻略生成 |
| P3: 进阶能力 | 2 周 | 向量检索 + SSE 流式 + RAG | ✅ 智能问答 |
| P4: 生产加固 | 1 周 | 监控 + 限流 + 安全 | ✅ Dashboard |

---

## Phase 1：LLM 基础设施「让 AI 服务先跑起来」

**时间**：2 周 | **对应设计文档**：3.1 LLM客户端 + 3.2 提示词引擎 + 4.1 推荐接口

### 目标

搭建 Python AI 服务骨架，实现第一个端到端 AI 功能——智能推荐。

### 要实现的文件

```
ai-service/
├── main.py                          # FastAPI 入口
├── requirements.txt                 # fastapi, uvicorn, openai, httpx, pydantic
├── .env                             # LLM_API_KEY
├── config/settings.py               # 配置加载
│
├── app/
│   ├── api/v1/
│   │   └── recommend.py             # POST /ai/recommend
│   │
│   ├── core/
│   │   ├── llm_client.py            # LLM 客户端（统一接口）
│   │   └── prompt_engine.py         # 提示词引擎（模板加载+变量填充）
│   │
│   ├── services/
│   │   └── recommendation_service.py # 推荐业务逻辑
│   │
│   ├── models/
│   │   └── schemas.py               # Pydantic 请求/响应模型
│   │
│   └── prompts/
│       └── recommend.txt            # 推荐场景提示词模板
│
└── tests/
    └── test_recommend.py
```

### 关键代码片段（面试要能讲清楚）

**LLM Client — 多服务商切换**：
```python
# app/core/llm_client.py
class LLMClient:
    """统一封装 OpenAI / Claude / 硅基流动"""
    
    def __init__(self, provider: str, api_key: str, base_url: str):
        self.provider = provider
        # 所有服务商走 OpenAI-compatible API
        self.client = AsyncOpenAI(api_key=api_key, base_url=base_url)
    
    async def chat(self, messages: list[dict], **kwargs) -> str:
        """带重试的 LLM 调用"""
        for attempt in range(3):
            try:
                resp = await self.client.chat.completions.create(
                    model=self._get_model(),
                    messages=messages,
                    temperature=0.7,
                    timeout=30,
                )
                return resp.choices[0].message.content
            except Exception as e:
                if attempt == 2: raise
                await asyncio.sleep(2 ** attempt)  # 指数退避
```

**Prompt Engine — 模板与代码分离**：
```python
# app/core/prompt_engine.py
class PromptEngine:
    """从 txt 文件加载模板，动态填充参数"""
    
    def __init__(self, templates_dir: str = "app/prompts"):
        self.templates_dir = Path(templates_dir)
        self._cache: dict[str, str] = {}
    
    def render(self, template_name: str, **kwargs) -> str:
        """加载模板 → 填充变量 → 返回完整 prompt"""
        if template_name not in self._cache:
            path = self.templates_dir / f"{template_name}.txt"
            self._cache[template_name] = path.read_text(encoding="utf-8")
        template = self._cache[template_name]
        return template.format(**kwargs)
```

### 降级策略

```python
# 当 LLM 不可用时，返回预设推荐
FALLBACK_RECOMMENDATIONS = [
    {"title": "成都三日美食之旅", "destinations": ["宽窄巷子","锦里","春熙路"], ...},
    {"title": "杭州西湖经典路线", "destinations": ["断桥残雪","苏堤","雷峰塔"], ...},
]

async def recommend_with_fallback(req, llm_client, prompt_engine):
    try:
        prompt = prompt_engine.render("recommend", query=req.query, ...)
        result = await llm_client.chat([{"role":"user","content":prompt}])
        return parse_json(result)
    except Exception:
        logger.warning("LLM unavailable, using fallback")
        return FALLBACK_RECOMMENDATIONS
```

### 面试话术

> "Phase 1 我搭建了 Python AI 服务的骨架。做了三个关键设计：
> 1. **LLM Client 抽象层**——所有服务商走 OpenAI-compatible API，切换只需要改配置文件
> 2. **提示词模板化**——prompt 存 txt 文件，运营人员可以直接改，不用动代码
> 3. **降级兜底**——指数退避重试 3 次，全失败后返回预设推荐数据"

### 可演示内容

```bash
# 启动服务
uvicorn main:app --port 8000

# 调推荐接口
curl -X POST http://localhost:8000/api/v1/ai/recommend \
  -H "Content-Type: application/json" \
  -d '{"query":"周末想去海边","days":2,"budget":"medium","interests":["beach","food"]}'

# 返回
{"code":200, "data":{"recommendations":[...]}}

# 模拟 LLM 挂了：停掉 API Key → 仍返回预设推荐
```

---

## Phase 2：Agent 核心「从"调 API"到"自主推理"」

**时间**：3 周 | **对应设计文档**：3.3 记忆系统 + 3.4 工具调用框架 + 4.2 攻略生成 + 4.4 智能客服

### 目标

这是面试最有区分度的阶段——让 Agent 能**自主决定调用什么工具、记住上下文、多步骤完成任务**。

### 新增文件

```
app/
├── api/v1/
│   ├── guide.py                     # POST /ai/guide  (攻略生成)
│   └── chat.py                      # POST /ai/chat   (智能客服)
│
├── core/
│   ├── tool_registry.py             # 工具注册中心 ← 核心！
│   └── result_parser.py             # LLM 响应解析
│
├── services/
│   ├── planning_service.py          # 行程规划
│   └── chat_service.py              # 智能客服
│
├── tools/
│   ├── map_tool.py                  # 地图搜索
│   ├── poi_tool.py                  # POI 查询
│   └── weather_tool.py              # 天气查询
│
├── memory/
│   ├── short_term_memory.py         # 会话上下文 (Redis)
│   └── preference_store.py          # 用户偏好 (MySQL)
│
└── prompts/
    ├── guide.txt                    # 攻略生成 prompt
    └── chat.txt                     # 客服对话 prompt
```

### 核心实现：Tool Registry + Function Calling

```python
# app/core/tool_registry.py
from typing import Any, Callable
from dataclasses import dataclass

@dataclass
class ToolParam:
    name: str
    type: str        # "string" | "number" | "boolean"
    description: str
    required: bool = True

class Tool:
    """可被 Agent 调用的工具"""
    def __init__(self, name: str, description: str, 
                 parameters: list[ToolParam], 
                 handler: Callable):
        self.name = name
        self.description = description
        self.parameters = parameters
        self.handler = handler
    
    def to_openai_schema(self) -> dict:
        """转为 OpenAI Function Calling 格式"""
        return {
            "type": "function",
            "function": {
                "name": self.name,
                "description": self.description,
                "parameters": {
                    "type": "object",
                    "properties": {
                        p.name: {"type": p.type, "description": p.description}
                        for p in self.parameters
                    },
                    "required": [p.name for p in self.parameters if p.required],
                }
            }
        }

class ToolRegistry:
    """工具注册中心 —— Agent 的能力边界"""
    
    def __init__(self):
        self._tools: dict[str, Tool] = {}
    
    def register(self, tool: Tool):
        self._tools[tool.name] = tool
    
    def get_schemas(self) -> list[dict]:
        return [t.to_openai_schema() for t in self._tools.values()]
    
    async def execute(self, name: str, params: dict) -> Any:
        if name not in self._tools:
            raise ValueError(f"Unknown tool: {name}")
        return await self._tools[name].handler(**params)

# ── 注册工具 ──
registry = ToolRegistry()
registry.register(Tool(
    name="search_attractions",
    description="搜索指定城市的旅游景点",
    parameters=[
        ToolParam("city", "string", "城市名称"),
        ToolParam("category", "string", "分类: nature/culture/food/shopping", required=False),
    ],
    handler=poi_service.search,
))
registry.register(Tool(
    name="query_weather",
    description="查询指定城市未来天气",
    parameters=[
        ToolParam("city", "string", "城市名称"),
        ToolParam("days", "number", "查询天数", required=False),
    ],
    handler=weather_service.query,
))
```

### Agent 循环 —— 多步推理

```python
# app/services/planning_service.py
async def generate_guide(user_request: str, tools: ToolRegistry, 
                          llm: LLMClient, memory: ShortTermMemory):
    """Agent 主循环：思考 → 调工具 → 再思考 → 输出"""
    
    messages = [{"role": "system", "content": GUIDE_SYSTEM_PROMPT}]
    messages.append({"role": "user", "content": user_request})
    
    # 加载会话历史
    history = await memory.get_session(user_request.session_id)
    messages = history + messages
    
    for _ in range(5):  # 最多 5 轮推理
        response = await llm.chat(
            messages=messages,
            tools=tools.get_schemas(),  # ← OpenAI Function Calling
        )
        
        if response.tool_calls:
            # LLM 决定调用工具
            for tc in response.tool_calls:
                result = await tools.execute(tc.name, tc.params)
                messages.append({"role": "tool", "content": str(result), 
                                 "tool_call_id": tc.id})
        else:
            # LLM 认为可以给出最终答案
            await memory.save_session(user_request.session_id, messages)
            return parse_guide(response.content)
    
    raise Exception("Agent exceeded max reasoning steps")
```

### 面试话术

> "Phase 2 我实现了 Agent 的核心能力——工具调用和记忆系统。
> 
> **Tool Registry** 参考了 OpenAI Function Calling 的设计：每个工具声明自己的名称、描述、参数 schema，LLM 根据用户意图自主决定调用哪个工具、传什么参数。比如用户说'生成杭州三日游攻略'，Agent 会先调 `search_attractions("杭州")` 拿景点列表，再调 `query_weather("杭州")` 查天气，最后综合信息生成攻略。
> 
> **分层记忆**：短期记忆用 Redis 存会话上下文（多轮对话不丢），长期偏好存 MySQL（下次打开 App 自动加载历史偏好）。
> 
> 这个架构的价值在于**可扩展**——新增一个工具只需要注册，不用改 Agent 核心逻辑。"

### 可演示内容

```bash
# 攻略生成——Agent 自主调用工具链
curl -X POST http://localhost:8000/api/v1/ai/guide \
  -d '{"city":"杭州","days":3,"interests":["culture","food"]}'

# Agent 后台实际执行：
# Step 1: search_attractions("杭州", "culture") → [西湖,灵隐寺,雷峰塔...]
# Step 2: query_weather("杭州", 3) → {晴,25-32°C}
# Step 3: LLM 综合信息 → 生成完整攻略 JSON

# 返回
{"code":200, "data":{
  "city":"杭州", "days":3,
  "daily_itinerary":[
    {"day":1, "title":"西湖经典一日", 
     "spots":[{"name":"断桥残雪","duration":"1h","tips":"早上去人少"},...],
     "food":"楼外楼 - 西湖醋鱼", "weather_note":"晴 28°C 适合户外"}
  ],...
}}
```

---

## Phase 3：进阶能力「向量检索 + 流式 + RAG」

**时间**：2 周 | **对应设计文档**：3.3 长期记忆(向量) + 4.4 智能客服增强

### 目标

加入向量检索和流式输出——这两个是面试中最容易被追问的技术点。

### 新增/修改

```
app/
├── core/
│   └── embedding_client.py          # Embedding 生成
│
├── memory/
│   └── long_term_memory.py          # 向量存储 (Redis Stack)
│
├── api/v1/
│   └── chat.py                      # 改为 SSE 流式输出
│
├── knowledge/                       # RAG 知识库
│   ├── city_guides/                 # 城市攻略文档
│   └── travel_tips.txt              # 旅行小贴士
│
└── prompts/
    └── rag_chat.txt                 # RAG 增强对话 prompt
```

### 关键实现

**向量检索**：
```python
# app/memory/long_term_memory.py
class LongTermMemory:
    """用向量相似度做个性化推荐"""
    
    def __init__(self, redis_client):
        self.redis = redis_client
    
    async def store_user_profile(self, user_id: str, 
                                  interests: list[str], 
                                  embedder: EmbeddingClient):
        """用户兴趣 → embedding → Redis 向量存储"""
        text = " ".join(interests)
        vec = await embedder.embed(text)
        await self.redis.ft("user_idx").add(
            redis.commands.search.document.Document(
                id=f"user:{user_id}",
                vector=vec,
                payload={"interests": ",".join(interests)}
            )
        )
    
    async def find_similar_users(self, user_id: str, top_k: int = 5):
        """找兴趣最相似的 Top-K 用户 → 协同过滤推荐"""
        vec = await self.redis.hget(f"user_vec:{user_id}", "vector")
        results = await self.redis.ft("user_idx").search(
            query_vector=vec, top_k=top_k, return_fields=["interests"]
        )
        return [r.payload for r in results.docs]
```

**SSE 流式输出**：
```python
# app/api/v1/chat.py
@router.post("/ai/chat")
async def chat_stream(req: ChatRequest):
    """SSE 流式返回——打字机效果"""
    return StreamingResponse(
        generate_chat_stream(req),
        media_type="text/event-stream",
    )

async def generate_chat_stream(req: ChatRequest):
    messages = build_messages(req)
    stream = await llm_client.chat_stream(messages)
    
    full_response = ""
    async for chunk in stream:
        if chunk.content:
            full_response += chunk.content
            yield f"data: {json.dumps({'content': chunk.content})}\n\n"
    
    yield f"data: {json.dumps({'done': True, 'full': full_response})}\n\n"
```

### 面试话术

> "Phase 3 我做了两个进阶能力：
> 
> **向量相似推荐**：把用户兴趣标签做 embedding，存到 Redis Stack。当需要推荐时，向量检索找到最相似的用户，做协同过滤。这个方案比传统协同过滤好在——不需要离线跑矩阵分解，实时写入、实时检索。
> 
> **SSE 流式输出**：前端用 EventSource 接收，实现类似 ChatGPT 的打字效果。关键设计是后端用 `asyncio` 异步生成器，每收到一个 token 就推给前端，不等完整响应。"

---

## Phase 4：生产加固「监控 + 安全 + 部署」

**时间**：1 周 | **对应设计文档**：第 6 章监控 + 第 7 章安全

### 目标

让面试官看到你有生产意识——不只是"能跑"，还要"跑得稳"。

### 内容

| 项目 | 实现 | 面试能说什么 |
|------|------|------------|
| OpenTelemetry 追踪 | 在 LLM Client 和 Tool 调用处加 span | "全链路追踪，每个工具调用的耗时和参数都能看到" |
| Token 用量监控 | Counter 统计每次 LLM 调用 token 数 | "做了 token 预算监控，避免恶意请求烧钱" |
| 限流 | 推荐 100/min, 聊天 50/min | "用 Redis sliding window 做 API 限流" |
| 内容安全 | 敏感词过滤 + 输出 JSON schema 校验 | "输入做敏感词过滤，输出做 schema 校验防止 prompt injection" |
| Docker Compose | Go + Python + Redis + MySQL 一键启动 | "docker-compose up 就能跑全栈" |

### 面试话术

> "做项目不能只考虑 happy path。我加了 OpenTelemetry 全链路追踪、token 用量监控、API 限流、内容安全过滤。用 Docker Compose 打包部署，面试官可以直接跑起来看效果。"

---

## 时间线总览

```
Week 1-2   ████████  Phase 1   LLM Client + Prompt Engine + 推荐接口
Week 3-5   ████████  Phase 2   Tool Registry + 记忆系统 + 攻略生成
Week 6-7   ████████  Phase 3   向量检索 + SSE + RAG
Week 8     ████████  Phase 4   监控 + 部署 + 文档

总计 8 周，每周约 10-15 小时
```

## 每个 Phase 结束的交付物清单

| Phase | 代码 | 文档 | 面试准备 |
|-------|------|------|---------|
| P1 | 6 个 Python 文件 | API 文档 + curl 示例 | 能讲清楚 LLM Client 抽象和降级策略 |
| P2 | 12 个 Python 文件 | 架构图更新 + Tool 说明 | 能讲清楚 Function Calling 原理和记忆分层 |
| P3 | 6 个 Python 文件 | 向量检索原理 + SSE 协议 | 能画 embedding→检索→推荐的流程图 |
| P4 | Dockerfile + compose | 部署文档 + 监控截图 | 能讲生产环境的监控和限流方案 |

---

*计划版本：v1.0 | 基于 AI_AGENT_DESIGN.md v1.0*
