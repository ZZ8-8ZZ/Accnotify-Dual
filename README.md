# Accnotify

Accnotify 是基于 **[trah01/Accnotify](https://github.com/trah01/Accnotify)** 二次开发的极简、纯净、安全私有化即时通知推送工具。**部署一套后端，即可同时支持 iOS 端 Bark 与 Android 端 Accnotify 双平台推送**，无需分别搭建服务，一套环境统一管理。

通过集成 **APNs (Apple Push Notification Service)** 实现对 **Bark** 客户端的完整后端兼容，同时保留原项目 Android 端稳定推送与超强保活能力，让多设备用户只需维护一个私有推送服务。

## 🌟 核心特性

- **Bark 完美兼容**：
  - **全接口支持**：支持 Bark 所有的推送接口，包括 `/:key/:body`、`/:key/:title/:body` 以及 `/:key/:title/:subtitle/:body`。
  - **全参数支持**：完整支持 `sound` (铃声)、`badge` (角标)、`group` (分组)、`icon` (自定义图标) 和 `url` (点击跳转) 等参数。
  - **自动验证**：完美支持 Bark 客户端添加服务器时的 `/ping` 自动验证。
- **Android 超强保活**：针对 Android 端，采用辅助功能（Accessibility Service）作为保活锚点，确保在各种国产 ROM（ColorOS, HyperOS, Flyme 等）下消息不延迟。
- **内容加密传输**：Android 端支持 RSA+AES 端到端加密；iOS 端沿用 Bark 原生加密逻辑，全平台传输安全可靠。
- **私有化部署**：支持 Docker 一键部署，数据完全掌握在自己手中。
- **Webhook 集成**：原生支持 GitHub, GitLab, Docker Hub, Gitea 等主流平台的 Webhook 推送。
- **离线存储**：支持推送历史记录，方便随时查阅。
- **无感运行**：无需 Root，不依赖 Xposed，不修改系统，安装即可使用。

## 🚀 快速开始

### 1. 部署服务器 (Docker)

确保您的服务器已安装 Docker 和 Docker Compose。

```bash
# 克隆项目
git clone https://github.com/ZZ8-8ZZ/Accnotify-Dual.git
# 进入 server 目录
cd server
# (可选) 手动构建镜像
docker-compose build
# 启动服务
docker-compose up -d
# 查看日志
docker-compose logs -f
```

部署成功后，访问 `http://your-server-ip:8080`，如果看到以下内容则表示服务正常运行：

```json
{
  "code": 200,
  "data": {
    "version": "1.0.0"
  },
  "message": "pong",
  "status": "ok",
  "timestamp": 1234567890
}
```

### 2. 配置 iOS 端 (Bark)

1. 在 App Store 下载 [Bark](https://apps.apple.com/app/id1403753865)
2. 打开 Bark，点击右上角 "+" 添加私有服务器
3. 输入您的服务器地址（例如 `https://push.example.com`）
4. 验证通过后，您将获得一个专属的推送 Key

### 3. 配置 Android 端 (Accnotify)

1. 下载安装 [Accnotify](https://github.com/trah01/Accnotify/releases) APK
2. 输入您的服务器地址（例如 `https://push.example.com`）
3. 开启「辅助功能保活」
4. 获得推送 Key，开始享受无感推送

## 📝 推送示例 (Bark 兼容格式)

**简单的 GET 请求：**

```bash
curl "https://your-server.com/your-key/这是消息内容"
```

**带标题、分组和铃声的请求：**

```bash
curl "https://your-server.com/your-key/这是标题/这是内容?group=测试&sound=minuet"
```

## ⚓ Webhook 集成

本项目支持多种 Webhook 格式，方便将第三方服务的通知转发至您的设备。

- **通用 Webhook (POST JSON)**: `https://your-server.com/webhook/your-key`
- **GitHub**: `https://your-server.com/webhook/your-key/github`
- **GitLab**: `https://your-server.com/webhook/your-key/gitlab`
- **Docker Hub**: `https://your-server.com/webhook/your-key/docker`
- **Gitea**: `https://your-server.com/webhook/your-key/gitea`

## 🔗 相关链接

- **Bark iOS 客户端：**: <https://github.com/Finb/Bark>
- **原项目 Accnotify：**: <https://github.com/trah01/Accnotify> (本项目二开核心基础)

## 🤝 致谢

感谢 [Bark](https://github.com/Finb/Bark) 项目提供的优秀 iOS 推送方案，本项目基于 trah01/Accnotify 扩展，实现 Bark 协议全兼容与 iOS + Android 双平台统一推送，致力于打造更易用的私有化推送服务。

## ❓ 常见问题（FAQ）

1. **Q：Bark 提示验证失败？**
   A：检查服务器端口是否通、防火墙是否放行（默认 8080）。虽然支持 HTTP，但建议使用 HTTPS 以确保推送内容的安全性（iOS Bark 客户端支持自定义 HTTP 服务器地址）。
2. **Q：Android 收不到推送 / 延迟？**
   A：开启「辅助功能保活」，关闭电池优化，锁定后台。
3. **Q：Docker 启动失败？**
   A：确认端口 8080 未被占用，重启 Docker 服务。
4. **Q：推送 Key 丢失怎么办？**
   A：重装 App 或清除数据重新绑定服务器。
5. **Q：支持自定义端口吗？**
   A：支持，修改 `docker-compose.yml` 中端口映射即可。

***

*如果您觉得本项目对您有帮助，欢迎给一个 Star！*
