# Local File Registry Configuration

本地文件注册中心配置文件说明。

## 文件结构

本地文件注册中心使用 JSON 格式存储服务注册信息，文件结构如下：

### 根级别字段

- `services`: 服务实例映射表，键为服务名称，值为该服务的实例数组
- `version`: 注册文件格式版本
- `updated`: 最后更新时间戳（Unix 时间戳）

### 服务实例字段 (ServiceInstance)

每个服务实例包含以下字段：

- `id`: 服务实例唯一标识符
- `name`: 服务名称
- `version`: 服务版本
- `metadata`: 服务元数据，键值对形式存储额外信息
  - `weight`: 负载均衡权重
  - `region`: 部署区域
  - `env`: 环境标识（如 production, staging, development）
  - 其他自定义元数据
- `endpoints`: 服务端点列表，支持 HTTP 和 gRPC 协议
- `timestamp`: 服务实例注册时间戳

## 示例配置

参考 `registry.json` 文件，包含了三个服务的示例配置：

1. **user-service**: 用户服务，包含两个实例
2. **order-service**: 订单服务，包含一个实例
3. **payment-service**: 支付服务，包含一个实例

## 使用方式

```go
// 创建本地文件注册中心
registry, err := local.New("/path/to/registry.json")
if err != nil {
    log.Fatal(err)
}

// 注册服务实例
service := &registry.ServiceInstance{
    ID:       "my-service-001",
    Name:     "my-service",
    Version:  "v1.0.0",
    Metadata: map[string]string{
        "weight": "100",
        "env":    "production",
    },
    Endpoints: []string{
        "http://localhost:8080",
        "grpc://localhost:9090",
    },
}

err = registry.Register(context.Background(), service)
if err != nil {
    log.Fatal(err)
}
```

## 注意事项

1. 文件会在首次使用时自动创建
2. 服务实例的 `timestamp` 字段用于跟踪注册时间
3. `metadata` 字段可以存储任意键值对，用于服务发现时的过滤和路由
4. `endpoints` 支持多种协议，格式为 `protocol://host:port`
5. 同一服务可以有多个实例，用于负载均衡