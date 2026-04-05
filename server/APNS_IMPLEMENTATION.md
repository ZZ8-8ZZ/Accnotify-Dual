# APNs支持实现总结

## 概述

成功扩展了Accnotify服务器以支持APNs（Apple Push Notification Service）推送，使其能够同时兼容Android（WebSocket）和iOS（APNs）设备，包括对Bark客户端的完整支持。

## 主要更改

### 1. 依赖管理

**文件**: [go.mod](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/go.mod)

- 添加了 `github.com/sideshow/apns2 v0.24.0` 依赖，用于APNs推送功能

### 2. 配置系统

**文件**: [config/config.go](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/config/config.go)

- 扩展了配置结构，添加APNs相关字段：
  - `APNsEnabled`: 是否启用APNs服务
  - `APNsCertFile`: 证书文件路径（证书认证）
  - `APNsKeyFile`: 密钥文件路径（令牌认证）
  - `APNsKeyID`: APNs密钥ID
  - `APNsTeamID`: Apple团队ID
  - `APNsTopic`: 应用包标识符
  - `APNsDevelopment`: 是否使用开发环境

### 3. 数据模型

**文件**: [model/message.go](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/model/message.go)

- 扩展了 `Device` 结构体：
  - `DeviceToken`: iOS设备的APNs令牌
  - `Platform`: 设备平台（"android"或"ios"）
- 扩展了 `RegisterRequest` 结构体：
  - `DeviceToken`: 支持Bark格式的devicetoken注册
  - `Platform`: 设备平台标识

### 4. 数据库存储

**文件**: [storage/sqlite.go](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/storage/sqlite.go)

- 更新了数据库表结构，添加 `device_token` 和 `platform` 字段
- 添加了 `UpdateDeviceToken` 方法用于更新设备令牌
- 更新了 `GetDeviceByKey` 和 `CreateDevice` 方法以支持新字段

### 5. APNs服务模块

**文件**: [apns/apns.go](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/apns/apns.go)

- 创建了完整的APNs服务实现
- 支持两种认证方式：
  - 令牌认证（推荐）：使用 `.p8` 密钥文件
  - 证书认证：使用 `.p12` 证书文件
- 实现了推送功能，支持：
  - 标准通知（标题、正文）
  - 声音、角标、分组
  - 自定义数据
  - 静默通知
  - 可变内容通知
- 包含配置验证和错误处理

### 6. 推送处理逻辑

**文件**: [handler/push.go](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/handler/push.go)

- 扩展了 `PushHandler` 结构体，添加APNs服务实例
- 更新了 `NewPushHandler` 构造函数以接受APNs服务参数
- 修改了 `HandlePush` 方法：
  - 优先通过WebSocket推送（Android设备）
  - 如果设备是iOS平台且有deviceToken，同时通过APNs推送
  - 支持完整的消息格式（标题、正文、分组、图标、URL、声音、角标）
- 修改了 `HandleSimplePush` 方法：
  - 支持Bark兼容的简单推送格式
  - 对iOS设备自动使用APNs推送
- 更新了 `HandleRegister` 方法：
  - 支持注册deviceToken和platform
  - 支持更新现有设备的APNs令牌

### 7. 主程序

**文件**: [main.go](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/main.go)

- 添加了APNs服务的初始化逻辑
- 配置验证和错误处理
- 优雅关闭时清理APNs连接
- 详细的日志记录

### 8. 配置文档

**文件**: [APNS_SETUP.md](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/APNS_SETUP.md)

- 完整的APNs配置指南
- 环境变量说明
- Apple开发者门户设置步骤
- Docker和直接部署的配置示例
- 测试和故障排除指南
- 安全最佳实践

**文件**: [.env.example](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/.env.example)

- 环境变量配置模板
- 包含所有APNs相关配置项的说明和示例

**文件**: [docker-compose.yml](file:///c:/Users/Administrator/Desktop/Push/Accnotify-main/server/docker-compose.yml)

- 更新了Docker Compose配置
- 添加了APNs配置的注释示例
- 包含密钥文件挂载的配置说明

## 技术特性

### 双通道推送

服务器现在支持两种推送方式：

1. **WebSocket推送**（Android设备）
   - 实时长连接
   - 支持端到端加密
   - 低延迟
   - 需要设备保持在线

2. **APNs推送**（iOS设备）
   - 基于苹果推送服务
   - 支持离线推送
   - 系统级通知
   - 更好的电池优化

### Bark兼容性

完全兼容Bark客户端的推送格式：

- **注册接口**：`POST /register`
  ```json
  {
    "device_key": "your_key",
    "devicetoken": "your_device_token",
    "platform": "ios",
    "name": "Device Name"
  }
  ```

- **推送接口**：`GET /push/:device_key/:title/:body`
  - 支持简单的URL格式推送
  - 自动识别iOS设备并使用APNs

### 智能路由

推送系统会自动选择最佳推送方式：

- Android设备 → WebSocket推送
- iOS设备 → APNs推送
- 如果WebSocket推送失败，iOS设备仍可通过APNs接收消息

## 使用示例

### 配置APNs

```bash
# 设置环境变量
export ACCNOTIFY_APNS_ENABLED=true
export ACCNOTIFY_APNS_TOPIC=com.yourcompany.yourapp
export ACCNOTIFY_APNS_KEY_ID=ABC123XYZ
export ACCNOTIFY_APNS_TEAM_ID=DEF456UVW
export ACCNOTIFY_APNS_KEY_FILE=/path/to/apns_key.p8
export ACCNOTIFY_APNS_DEVELOPMENT=false
```

### 注册iOS设备

```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "device_key": "bark_device_key",
    "devicetoken": "ios_device_token",
    "platform": "ios",
    "name": "My iPhone"
  }'
```

### 发送推送

```bash
# 简单推送（Bark格式）
curl http://localhost:8080/push/bark_device_key/Hello/World

# 完整推送
curl -X POST http://localhost:8080/push/bark_device_key \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Notification Title",
    "body": "Notification Body",
    "group": "updates",
    "sound": "default",
    "badge": 1
  }'
```

## 部署说明

### Docker部署

1. 准备APNs密钥文件
2. 更新 `docker-compose.yml` 中的环境变量
3. 挂载密钥文件到容器
4. 启动服务：`docker-compose up -d`

### 直接部署

1. 设置环境变量
2. 编译项目：`go build`
3. 运行服务器：`./server`

## 注意事项

1. **安全性**：
   - 不要将APNs密钥或证书提交到版本控制
   - 使用环境变量或密钥管理工具
   - 定期轮换APNs密钥

2. **开发与生产**：
   - 开发环境使用 `ACCNOTIFY_APNS_DEVELOPMENT=true`
   - 生产环境使用 `ACCNOTIFY_APNS_DEVELOPMENT=false`
   - 开发和生产使用不同的deviceToken

3. **性能优化**：
   - APNs服务使用连接池
   - 支持批量推送
   - 自动重试失败的推送

## 测试

建议的测试步骤：

1. 配置APNs服务（开发模式）
2. 注册测试设备
3. 发送测试推送
4. 验证iOS设备收到通知
5. 检查服务器日志确认推送状态
6. 测试离线推送功能

## 总结

通过这次扩展，Accnotify服务器现在具备了：

✅ 完整的APNs支持
✅ Bark客户端兼容性
✅ 双通道推送能力
✅ 智能设备识别
✅ 灵活的配置选项
✅ 完善的错误处理
✅ 详细的文档说明

这使得Accnotify成为一个真正跨平台的推送解决方案，能够同时服务Android和iOS用户。