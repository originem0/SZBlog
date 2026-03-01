# SZBlog

基于 Hugo + PaperMod 的个人博客，部署在 Ubuntu 服务器上，通过 GitHub Webhook 实现自动部署。

## 快速开始

```
本地: 写 Markdown → git push
         ↓
GitHub: 存储源码（博客安全网）
         ↓ (Webhook 通知)
服务器: git fetch + reset → hugo build → Nginx 托管静态文件
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
- [x] GitHub Webhook 自动部署（Go 服务，HMAC 签名验证）
- [x] 归档页面（按年月分组）
- [x] 全文搜索（PaperMod 内置，基于 JSON 索引）
- [x] 关于页面
- [x] giscus 评论系统（基于 GitHub Discussions）
- [x] 阅读统计（Go + SQLite 聚合计数，Nginx 反向代理）
- [x] Nginx gzip 压缩 + 静态资源缓存
- [x] 视觉优化：GitHub Markdown 风格（Noto Sans SC / Inter / JetBrains Mono）
- [x] 视觉优化：排版调整（行高、间距、链接样式、代码块圆角）
- [x] 视觉优化：列表页扁平化（去卡片，底线分隔）+ 归档页紧凑化
- [x] 视觉优化：Paginav 重设计、导航栏过渡、标签/面包屑/Footer 交互优化
- [x] 亮色/暗色模式适配（GitHub Light / GitHub Dark 配色）
- [x] 代码高亮：亮色 GitHub 主题 + 暗色 Catppuccin Macchiato
- [x] 侧边栏分类导航（Hugo 模板动态生成，宽屏 sticky 浮动）
- [x] TOC 目录浮动（宽屏右侧 fixed，窄屏内联）
- [x] SEO：OpenGraph / Twitter Card meta（`env = "production"`）
- [x] Webhook 安全加固：secret 空拒绝、CORS 域名限定、path 校验、rate limit
- [x] CSS 模块化拆分（6 个文件，按职责分离）
- [x] Front Matter 统一 YAML 格式
- [x] 想法（thoughts）独立 section + 极简列表模板

## 项目结构

```
blog/
├── assets/css/extended/       # 自定义样式（PaperMod 按文件名字母序自动加载）
│   ├── 01-variables.css       # 亮色/暗色模式 CSS 变量
│   ├── 02-typography.css      # 字体栈、排版、标题大小
│   ├── 03-components.css      # 链接、blockquote、卡片、标签、导航等
│   ├── 04-code.css            # chroma 语法高亮
│   ├── 05-sidebar.css         # 侧边栏导航 + TOC 浮动
│   └── 06-responsive.css      # 响应式 media queries
├── content/
│   ├── posts/                 # 文章 Markdown 文件
│   ├── ai-daily/              # AI 日报板块
│   ├── thoughts/              # 短想法（极简列表模板）
│   ├── archives.md            # 归档页
│   ├── search.md              # 搜索页
│   └── about.md               # 关于页
├── layouts/
│   ├── partials/
│   │   ├── comments.html      # giscus 评论组件
│   │   ├── extend_head.html   # Google Fonts 字体加载
│   │   └── extend_footer.html # 阅读统计 JS + 侧边栏导航
│   └── thoughts/
│       └── list.html          # 想法列表专属模板（只显示日期+标题）
├── themes/PaperMod/           # 主题（Git Submodule）
├── webhook/                   # 自动部署 + 阅读统计 API（Go + SQLite）
├── hugo.toml                  # Hugo 配置
├── DEPLOY.md                  # 完整部署文档和故障排查
└── README.md
```

### 内容
- [ ] 写满 10 篇文章（第一个月目标）

### 优化
- [ ] 配置 CDN 加速静态资源
- [ ] 图片方案：接入对象存储（腾讯云 COS）作为图床，避免仓库膨胀
- [ ] 服务器安全加固：改 SSH 端口、关闭密码登录

### Webhook 改进
- [ ] webhook 构建失败时发通知（邮件或微信推送）
- [ ] 添加构建日志持久化，方便排查问题

## 运维常用命令

```bash
# 查看 webhook 构建日志（排查部署失败）
ssh ubuntu@111.230.5.121 "journalctl -u blog-webhook -n 50 --no-pager"

# 手动触发构建
ssh ubuntu@111.230.5.121 "cd /home/ubuntu/blog && git fetch origin && git reset --hard origin/main && hugo"

# 重启 webhook 服务
ssh ubuntu@111.230.5.121 "sudo systemctl restart blog-webhook"
```
