# Dynamic Registry Configuration for Kratos

这个模块为 Kratos 框架提供了动态注册中心配置功能，允许开发者根据不同的部署环境（如 Windows 开发环境使用 local，Kubernetes 生产环境使用 etcd）动态选择和配置注册中心。

## 特性

- 🔧 **动态配置**: 通过配置文件动态选择注册中心类型
- 🏗️ **工厂模式**: 使用工厂模式统一创建不同类型的注册中心
- 📋 **Proto配置**: 使用 Protocol Buffers 定义配置结构，符合 Kratos 规范
- 🔍 **配置验证**: 提供完整的配置验证功能
- 🌐 **多注册中心支持**: 支持 local、etcd、consul、nacos、kubernetes 等多种注册中心
- 🔄 **跨平台**: 支持 Windows、Linux、macOS 等多种操作系统

## 支持的注册中心类型

| 类型 | 描述 | 适用场景 |
|------|------|----------|
| `local` | 本地文件注册中心 | 开发环境、单机部署 |
| `etcd` | etcd 分布式注册中心 | 生产环境、微服务集群 |
| `consul` | Consul 服务发现 | 生产环境、多数据中心 |
| `nacos` | Nacos 注册中心 | 阿里云环境、Spring Cloud |
| `kubernetes` | Kubernetes 原生服务发现 | 容器化部署、K8s 环境 |

## 安装

```bash
go get github.com/cocosip/zero/contrib/registry
```

## 快速开始

### 1. 生成配置代码

首先，从 proto 文件生成 Go 配置代码：

```bash
# Windows (使用 PowerShell)
protoc --proto_path=. --proto_path=./third_party --go_out=paths=source_relative:. conf/conf.proto

# Linux/macOS (使用 Makefile)
make config
```

### 2. 配置文件示例

创建 `config.yaml` 配置文件：

```yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s

# 注册中心配置
registry:
  # 开发环境使用本地文件注册中心
  type: "local"
  local:
    file_path: "./registry.json"
```

### 3. 应用程序集成

```go
package main

import (
    "context"
    "github.com/go-kratos/kratos/v2"
    "github.com/go-kratos/kratos/v2/config"
    "github.com/go-kratos/kratos/v2/config/file"
    
    registry_factory "github.com/cocosip/zero/contrib/registry"
    registry_conf "github.com/cocosip/zero/contrib/registry/conf"
)

type Bootstrap struct {
    Server   *Server              `yaml:"server"`
    Registry *registry_conf.Registry `yaml:"registry"`
}

func main() {
    // 加载配置
    c := config.New(config.WithSource(file.NewSource("config.yaml")))
    defer c.Close()
    c.Load()
    
    var bc Bootstrap
    c.Scan(&bc)
    
    // 验证注册中心配置
    if err := registry_factory.ValidateConfig(bc.Registry); err != nil {
        panic(err)
    }
    
    // 创建注册中心
    factory := registry_factory.NewRegistryFactory()
    registrar, discovery, err := factory.CreateRegistry(context.Background(), bc.Registry)
    if err != nil {
        panic(err)
    }
    
    // 创建 Kratos 应用
    app := kratos.New(
        kratos.Name("my-service"),
        kratos.Registrar(registrar),
        // ... 其他配置
    )
    
    app.Run()
}
```

## 配置详解

### Local 注册中心配置

适用于开发环境和单机部署：

```yaml
registry:
  type: "local"
  local:
    file_path: "./registry.json"  # 注册文件路径
```

### Etcd 注册中心配置

适用于生产环境的分布式部署：

```yaml
registry:
  type: "etcd"
  etcd:
    endpoints:
      - "127.0.0.1:2379"
      - "127.0.0.1:2380"
    dial_timeout: "5s"
    username: ""          # 可选：认证用户名
    password: ""          # 可选：认证密码
    namespace: "/microservices"  # 可选：命名空间
```

### Consul 注册中心配置

适用于多数据中心部署：

```yaml
registry:
  type: "consul"
  consul:
    address: "127.0.0.1:8500"
    scheme: "http"        # http 或 https
    datacenter: "dc1"     # 数据中心
    token: ""             # 可选：ACL token
    namespace: ""         # 可选：命名空间
```

### Nacos 注册中心配置

适用于阿里云环境：

```yaml
registry:
  type: "nacos"
  nacos:
    server_configs:
      - ip_addr: "127.0.0.1"
        port: 8848
        context_path: "/nacos"
    client_config:
      namespace_id: "public"
      username: "nacos"
      password: "nacos"
      log_level: "info"
      log_dir: "./logs"
      cache_dir: "./cache"
    group: "DEFAULT_GROUP"
    cluster: "DEFAULT"
```

### Kubernetes 注册中心配置

适用于容器化部署：

```yaml
registry:
  type: "kubernetes"
  kubernetes:
    namespace: "default"           # K8s 命名空间
    kube_config: ""               # kubeconfig 文件路径（集群内部署时为空）
    in_cluster: true              # 是否在集群内运行
    label_selector: "app=microservice"  # 标签选择器
```

## 环境适配建议

### 开发环境

```yaml
# 适用于 Windows/macOS/Linux 开发环境
registry:
  type: "local"
  local:
    file_path: "./registry.json"
```

### 测试环境

```yaml
# 使用单节点 etcd
registry:
  type: "etcd"
  etcd:
    endpoints: ["etcd-test:2379"]
    dial_timeout: "3s"
```

### 生产环境

```yaml
# 使用高可用 etcd 集群
registry:
  type: "etcd"
  etcd:
    endpoints:
      - "etcd-1:2379"
      - "etcd-2:2379"
      - "etcd-3:2379"
    dial_timeout: "5s"
    namespace: "/prod/microservices"
```

### Kubernetes 环境

```yaml
# 使用 K8s 原生服务发现
registry:
  type: "kubernetes"
  kubernetes:
    namespace: "production"
    in_cluster: true
    label_selector: "tier=backend"
```

## API 文档

### RegistryFactory 接口

```go
type RegistryFactory interface {
    CreateRegistry(ctx context.Context, config *conf.Registry) (kratos_registry.Registrar, kratos_registry.Discovery, error)
}
```

### 主要函数

- `NewRegistryFactory()`: 创建注册中心工厂实例
- `ValidateConfig(config *conf.Registry)`: 验证注册中心配置
- `GetDefaultConfig()`: 获取默认配置（local 类型）

## 示例项目

查看 `example/` 目录下的完整示例：

- `main_with_config.go`: 使用动态配置的完整应用示例
- `config.yaml`: 配置文件示例
- `registry.json`: 本地注册中心数据文件示例

运行示例：

```bash
cd example
go run main_with_config.go -conf config.yaml
```

## 开发指南

### 添加新的注册中心类型

1. 在 `conf/conf.proto` 中添加新的配置消息类型
2. 重新生成 Go 代码：`protoc --proto_path=. --proto_path=./third_party --go_out=paths=source_relative:. conf/conf.proto`
3. 在 `factory.go` 中实现对应的创建方法
4. 更新配置验证逻辑

### 配置验证

所有配置都会通过 `ValidateConfig` 函数进行验证，确保必要的字段不为空。

### 错误处理

模块提供详细的错误信息，帮助开发者快速定位配置问题。

## 注意事项

1. **配置文件格式**: 支持 YAML、JSON 等格式
2. **环境变量**: 可以结合 Kratos 的环境变量功能动态切换配置
3. **安全性**: 生产环境建议使用认证和加密
4. **性能**: 不同注册中心的性能特性不同，请根据实际需求选择
5. **依赖管理**: 某些注册中心需要额外的客户端库依赖

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！