# 迹忆旅图 - 代码规范

## 项目结构

```
trailmemo/
├── cmd/                     # 应用入口
│   └── server/
│       └── main.go          # 主入口
├── internal/                # 内部模块（禁止被外部导入）
│   ├── config/              # 配置管理
│   │   ├── config.go        # 配置结构体定义
│   │   ├── database.go      # 数据库连接
│   │   └── redis.go         # Redis连接
│   ├── handler/             # HTTP处理层（控制器）
│   │   └── user.go          # 用户相关处理器
│   ├── service/             # 业务逻辑层
│   │   └── user.go          # 用户相关服务
│   ├── repository/           # 数据访问层
│   │   └── user.go          # 用户相关数据访问
│   ├── model/                # 数据模型
│   │   ├── model.go         # 核心数据模型
│   │   └── migration.go     # 数据库迁移
│   ├── middleware/           # 中间件
│   │   ├── auth.go          # JWT认证
│   │   ├── cors.go          # 跨域处理
│   │   └── logger.go        # 日志
│   └── utils/                # 工具函数
│       └── utils.go
├── pkg/                     # 公共包（可以被外部导入）
│   └── response/            # 统一响应格式
│       └── response.go
├── api/                     # API定义
│   └── docs/                # Swagger文档
├── configs/                 # 配置文件
│   └── config.yaml
└── migrations/               # 数据库迁移脚本
```

## 分层职责

### Handler（处理层）
- 接收HTTP请求参数
- 参数校验
- 调用Service层
- 返回统一响应格式
- **禁止**：业务逻辑直接处理、数据库操作

### Service（业务逻辑层）
- 业务逻辑处理
- 事务管理
- 业务规则校验
- **禁止**：直接处理HTTP请求、响应

### Repository（数据访问层）
- 数据库CRUD操作
- 复杂查询
- **禁止**：业务逻辑处理

## 命名规范

### 文件命名
- 使用下划线分隔：`user_service.go`
- 驼峰命名法：`userService.go`（仅在公开方法时使用）

### 结构体命名
- 大写字母开头（公开）：`type User struct{}`
- 小写字母开头（私有）：`type userRepository struct{}`

### 函数命名
- 公开方法：首字母大写
- 私有方法：首字母小写
- 测试函数：以 `Test` 开头

### 常量命名
- 全大写，下划线分隔：`const MAX_RETRY_COUNT = 3`
- 枚举类型：`type Status int; const (StatusActive Status = 1)`

## 代码风格

### 错误处理
```go
// 推荐
if err != nil {
    return err
}

// 不推荐
if err != nil {
    panic(err)
}
```

### Context使用
```go
func (h *UserHandler) GetUser(c *gin.Context) {
    ctx := c.Request.Context()
    user, err := h.userService.GetUser(ctx, userID)
}
```

### 日志记录
```go
logger.Info("message",
    zap.String("key1", value1),
    zap.Int("key2", value2),
)
```

## API响应格式

```json
{
    "code": 200,
    "message": "success",
    "data": {}
}
```

## 数据库表命名
- 使用下划线分隔：`checkin_records`
- 主键：`id`
- 创建时间：`created_at`
- 更新时间：`updated_at`
- 软删除：`deleted_at`

## 配置管理
- 所有敏感信息（密码、密钥）必须通过配置文件或环境变量
- 禁止硬编码敏感信息
