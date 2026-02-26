# SZBlog

基于 Hugo + PaperMod 的个人博客，部署在 Ubuntu 服务器上，通过 GitHub Webhook 实现自动部署。

## 快速开始

```
本地: 写 Markdown → git push
         ↓
GitHub: 存储源码（博客安全网）
         ↓ (Webhook 通知)
服务器: git pull → hugo build → Nginx 托管静态文件
```

### 写文章

```bash
hugo new content posts/文章名.md
# 编辑文章，把 draft 改为 false
git add content/posts/文章名.md
git commit -m "new post: 文章名"
git push
# 几秒后自动上线
```

### 本地预览

```bash
hugo server --buildDrafts
# 访问 http://localhost:1313
```

## 已完成功能

- [x] Hugo + PaperMod 博客框架
- [x] GitHub Webhook 自动部署（Go 服务）
- [x] 归档页面（按年月分组）
- [x] 全文搜索（PaperMod 内置，基于 JSON 索引）
- [x] 关于页面
- [x] giscus 评论系统（基于 GitHub Discussions）
- [x] 阅读统计（Go + SQLite，Nginx 反向代理）
- [x] Nginx gzip 压缩 + 静态资源缓存

## 项目结构

```
blog/
├── content/
│   ├── posts/             # 文章 Markdown 文件
│   ├── archives.md        # 归档页
│   ├── search.md          # 搜索页
│   └── about.md           # 关于页
├── layouts/partials/
│   ├── comments.html      # giscus 评论组件（覆盖主题）
│   └── extend_footer.html # 阅读统计 JS（覆盖主题）
├── themes/PaperMod/       # 主题（Git Submodule）
├── webhook/               # 自动部署 + 阅读统计服务（Go）
├── hugo.toml              # Hugo 配置
├── DEPLOY.md              # 完整部署文档和故障排查
└── README.md
```

## 线上地址

- 临时：http://111.230.5.121
- 正式：https://blog.xxx.xxx（域名备案中）

## TODO

### 域名与 HTTPS
- [ ] 域名备案完成后绑定 `blog.xxx.xxx`
- [ ] 配置 SSL 证书（Let's Encrypt 或腾讯云免费证书）
- [ ] Nginx 配置 HTTP → HTTPS 跳转
- [ ] 更新 hugo.toml 中的 baseURL

### 内容
- [ ] 写满 10 篇文章（第一个月目标）

### 优化
- [ ] 配置 CDN 加速静态资源
- [ ] 图片方案：接入对象存储（腾讯云 COS）作为图床，避免仓库膨胀
- [ ] 服务器安全加固：改 SSH 端口、关闭密码登录

### Webhook 改进
- [ ] webhook 构建失败时发通知（邮件或微信推送）
- [ ] 添加构建日志持久化，方便排查问题
