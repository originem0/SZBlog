# SZBlog 部署文档

## 架构

```
本地 Windows (写 Markdown)
    ↓ git push
GitHub (originem0/SZBlog)
    ↓ Webhook POST http://111.230.5.121:9000/webhook
服务器 Ubuntu (git pull → hugo build → Nginx 托管静态文件)
```

## 技术栈

- Hugo Extended v0.157.0（静态站点生成器）
- PaperMod 主题（Git Submodule）
- Nginx（Web 服务器 + 反向代理 + gzip 压缩）
- Go Webhook 服务（自动部署 + 阅读统计 API）
- SQLite（阅读计数存储）
- giscus（评论系统，基于 GitHub Discussions）
- GitHub（代码托管）

---

## 一、本地环境搭建

### 1. 安装工具

```bash
scoop install hugo-extended
git --version   # 确认 git 已安装
```

### 2. 创建 Hugo 项目

```bash
cd D:\Development\Apps
hugo new site blog
cd blog
git init
git submodule add https://github.com/adityatelange/hugo-PaperMod.git themes/PaperMod
```

### 3. 配置文件 hugo.toml

```toml
baseURL = 'http://111.230.5.121/'
languageCode = 'zh-cn'
title = 'My Blog'
theme = 'PaperMod'

[outputs]
  home = ["HTML", "RSS", "JSON"]

[params.author]
  name = "sharon"

[params]
  defaultTheme = "auto"
  ShowReadingTime = true
  ShowPostNavLinks = true
  ShowBreadCrumbs = true
  ShowShareButtons = false
  comments = true

[params.homeInfoParams]
  Title = "Welcome"
  Content = "个人博客"

[[menus.main]]
  name = "文章"
  url = "/posts/"
  weight = 1

[[menus.main]]
  name = "归档"
  url = "/archives/"
  weight = 2

[[menus.main]]
  name = "搜索"
  url = "/search/"
  weight = 3

[[menus.main]]
  name = "关于"
  url = "/about/"
  weight = 4
```

### 4. 创建 .gitignore

```
public/
resources/
.hugo_build.lock
```

### 5. 创建内容页面

```bash
# 归档页（PaperMod 内置布局，按年月分组展示所有文章）
# content/archives.md: layout: "archives"

# 搜索页（PaperMod 内置前端搜索，基于 JSON 索引）
# content/search.md: layout: "search"

# 关于页
# content/about.md
```

### 6. 推送到 GitHub

```bash
git add -A
git commit -m "init: Hugo blog with PaperMod theme"
git remote add origin git@github.com:originem0/SZBlog.git
git branch -M main
git push -u origin main
```

---

## 二、服务器部署

服务器：Ubuntu, 4C4G, IP 111.230.5.121

### 1. 安装 Hugo

**不用 `snap install hugo`，用二进制安装。** 原因：snap 安装的 hugo 在 systemd 服务（webhook）中执行时会报 `cannot set memlock limit` 权限错误，因为 snap 的沙箱机制限制了非交互环境下的系统调用。直接用二进制放到 `/usr/local/bin` 没有这个问题。

由于服务器访问 GitHub 慢，从本地下载后传上去：

```powershell
# 本地 Windows
curl -L -o D:\Development\Apps\hugo_linux.tar.gz https://github.com/gohugoio/hugo/releases/download/v0.157.0/hugo_extended_0.157.0_linux-amd64.tar.gz
scp D:\Development\Apps\hugo_linux.tar.gz ubuntu@111.230.5.121:/tmp/hugo.tar.gz
```

```bash
# 服务器
sudo tar -xzf /tmp/hugo.tar.gz -C /usr/local/bin hugo
rm /tmp/hugo.tar.gz
/usr/local/bin/hugo version
```

### 2. 克隆仓库并构建

```bash
cd /home/ubuntu
git clone https://github.com/originem0/SZBlog.git blog
cd blog
git submodule update --init --recursive --depth 1
hugo
```

如果 git submodule 卡住（服务器访问 GitHub 慢），从本地打包传：

```powershell
# 本地 Windows
cd D:\Development\Apps\blog
tar -czf papermod.tar.gz -C themes/PaperMod .
scp papermod.tar.gz ubuntu@111.230.5.121:/home/ubuntu/blog/
```

```bash
# 服务器
cd /home/ubuntu/blog
mkdir -p themes/PaperMod
tar -xzf papermod.tar.gz -C themes/PaperMod
rm papermod.tar.gz
/usr/local/bin/hugo
```

### 3. 配置 Nginx

```bash
sudo tee /etc/nginx/sites-enabled/blog > /dev/null << 'EOF'
server {
    listen 80 default_server;
    server_name 111.230.5.121;

    root /home/ubuntu/blog/public;
    index index.html;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript image/svg+xml;
    gzip_min_length 1024;
    gzip_vary on;

    location ~* \.(css|js|woff2?|ttf|eot|svg|ico|png|jpg|jpeg|gif|webp)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    location ~* \.html$ {
        add_header Cache-Control "no-cache";
    }

    # 反向代理阅读统计 API
    location /api/ {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
        try_files $uri $uri/ =404;
    }
}
EOF

sudo nginx -t && sudo systemctl reload nginx
```

**注意：** /home/ubuntu 目录权限需要允许 www-data 访问：

```bash
sudo chmod 755 /home/ubuntu
```

### 4. 部署 Webhook + 阅读统计服务

webhook 代码在仓库 `webhook/main.go` 中，同时包含自动部署和阅读统计 API。

**依赖：** 需要 Go 和 gcc（go-sqlite3 是 CGO 包）。

```bash
sudo snap install go --classic
sudo apt install -y gcc
```

```bash
# 编译
cd /home/ubuntu/blog/webhook
go mod tidy
CGO_ENABLED=1 go build -o /home/ubuntu/webhook-server .
```

创建 systemd 服务：

```bash
sudo tee /etc/systemd/system/blog-webhook.service > /dev/null << 'EOF'
[Unit]
Description=Blog Webhook Server
After=network.target

[Service]
User=ubuntu
Environment=WEBHOOK_SECRET=你的密码
ExecStart=/home/ubuntu/webhook-server
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable blog-webhook
sudo systemctl start blog-webhook
```

服务提供两组端点：
- `POST /webhook` — GitHub 推送时自动 git pull + hugo build
- `GET/POST /api/view?path=xxx` — 阅读计数（通过 Nginx 反向代理从 80 端口访问）

### 5. 配置 GitHub Webhook

GitHub 仓库 → Settings → Webhooks → Add webhook：

| 字段 | 值 |
|------|-----|
| Payload URL | `http://111.230.5.121:9000/webhook` |
| Content type | `application/json` |
| Secret | 与 systemd 中 WEBHOOK_SECRET 一致 |
| Events | Just the `push` event |

### 6. 配置 giscus 评论系统

前置步骤：
1. GitHub 仓库 Settings → Features → 启用 Discussions
2. 安装 giscus App：https://github.com/apps/giscus
3. 去 https://giscus.app 获取 data-repo-id 和 data-category-id

评论模板在 `layouts/partials/comments.html`，覆盖了 PaperMod 的空文件。

### 7. 腾讯云安全组

放行 TCP 9000 端口，来源限制为 GitHub IP 段：

```
192.30.252.0/22
185.199.108.0/22
140.82.112.0/20
143.55.64.0/20
```

阅读统计 API 通过 Nginx 反向代理走 80 端口，无需额外开放端口。

---

## 三、日常写文章流程

```bash
# 本地 Windows
cd D:\Development\Apps\blog

# 创建新文章
hugo new content posts/文章名.md

# 编辑文章：把 draft = true 改为 draft = false，写内容

# 本地预览（可选）
hugo server --buildDrafts

# 发布
git add content/posts/文章名.md
git commit -m "new post: 文章名"
git push

# 几秒后自动上线
```

---

## 四、域名备案完成后要改的地方

1. **hugo.toml** 中 `baseURL` 改为 `https://blog.你的域名/`

2. **Nginx 配置** 更新 server_name 并加 SSL：

```nginx
server {
    listen 80;
    server_name blog.你的域名;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl;
    server_name blog.你的域名;

    ssl_certificate     /etc/nginx/你的证书.crt;
    ssl_certificate_key /etc/nginx/你的密钥.key;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    root /home/ubuntu/blog/public;
    index index.html;

    gzip on;
    gzip_types text/plain text/css application/json application/javascript text/xml application/xml text/javascript image/svg+xml;
    gzip_min_length 1024;
    gzip_vary on;

    location ~* \.(css|js|woff2?|ttf|eot|svg|ico|png|jpg|jpeg|gif|webp)$ {
        expires 30d;
        add_header Cache-Control "public, immutable";
        access_log off;
    }

    location ~* \.html$ {
        add_header Cache-Control "no-cache";
    }

    location /api/ {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location / {
        try_files $uri $uri/ =404;
    }
}
```

3. **DNS 解析** 添加 A 记录：`blog` → `111.230.5.121`

---

## 五、故障自查步骤

### 推送后网站没更新

按顺序排查：

**第1步：GitHub Webhook 有没有发出去？**

GitHub 仓库 → Settings → Webhooks → 点进去看 Recent Deliveries。
- 绿色 = 发出且收到 200 响应
- 红色 = 发出但服务器返回错误或超时
- 没有记录 = Webhook 没触发，检查配置

**第2步：服务器 webhook 服务在跑吗？**

```bash
sudo systemctl status blog-webhook
```

如果 inactive/dead：

```bash
sudo systemctl start blog-webhook
```

**第3步：webhook 日志有没有报错？**

```bash
sudo journalctl -u blog-webhook --since "10 minutes ago" --no-pager
```

常见报错：
- `build failed: exit status 1` → hugo 构建失败，手动跑 `cd /home/ubuntu/blog && /usr/local/bin/hugo` 看具体错误
- `invalid signature` → WEBHOOK_SECRET 和 GitHub 上设的不一致
- 没有任何日志 → 请求没到服务器，检查腾讯云安全组 9000 端口

**第4步：hugo 构建有没有问题？**

```bash
cd /home/ubuntu/blog
git pull
/usr/local/bin/hugo
```

手动构建看有没有报错。

**第5步：Nginx 配置有没有问题？**

```bash
sudo nginx -t
sudo systemctl reload nginx
```

**第6步：文件权限**

```bash
ls -la /home/ubuntu/blog/public/index.html
namei -l /home/ubuntu/blog/public/index.html
```

确保 www-data 可以读取所有路径。

### 阅读统计不显示

1. 浏览器 F12 → Network，看 `/api/view` 请求有没有发出、状态码是什么
2. 确认 Nginx 配置中有 `/api/` 反向代理到 `127.0.0.1:9000`
3. 确认 webhook 服务在运行：`sudo systemctl status blog-webhook`
4. 检查 SQLite 数据库是否正常：`sqlite3 /home/ubuntu/blog-views.db "SELECT * FROM views LIMIT 5;"`

### 评论框不显示

1. 确认 `hugo.toml` 中 `comments = true`
2. 确认 `layouts/partials/comments.html` 中 data-repo-id 和 data-category-id 已填写
3. 确认 GitHub 仓库已启用 Discussions
4. 确认已安装 giscus App 到仓库
5. 浏览器 F12 → Console 看有没有 giscus 相关报错

### 其他常见问题

**文章发布了但页面上看不到：**
- 检查文章头部 `draft` 是否为 `false`
- 检查文章 `date` 是否是未来时间（Hugo 默认不发布未来日期的文章）

**主题样式丢失：**
- 检查 themes/PaperMod 目录是否存在且有内容
- `git submodule update --init --recursive`

**webhook 二进制更新后没生效：**
```bash
cd /home/ubuntu/blog/webhook
go mod tidy
CGO_ENABLED=1 go build -o /home/ubuntu/webhook-server .
sudo systemctl restart blog-webhook
```

**服务器磁盘空间不足（40G SSD）：**
```bash
df -h
# 清理 snap 缓存
sudo sh -c 'rm -rf /var/lib/snapd/cache/*'
# 清理旧日志
sudo journalctl --vacuum-size=100M
```

---

## 六、关键文件位置

| 文件 | 位置 |
|------|------|
| 本地项目 | `D:\Development\Apps\blog` |
| Hugo 配置 | `hugo.toml` |
| 文章目录 | `content/posts/` |
| 归档页 | `content/archives.md` |
| 搜索页 | `content/search.md` |
| 关于页 | `content/about.md` |
| 评论模板 | `layouts/partials/comments.html` |
| 阅读统计 JS | `layouts/partials/extend_footer.html` |
| 主题 | `themes/PaperMod/` (Git Submodule) |
| 服务器博客目录 | `/home/ubuntu/blog` |
| 服务器 Hugo 二进制 | `/usr/local/bin/hugo` |
| Webhook 二进制 | `/home/ubuntu/webhook-server` |
| Webhook 源码 | `webhook/main.go` |
| Webhook systemd 配置 | `/etc/systemd/system/blog-webhook.service` |
| Nginx 博客配置 | `/etc/nginx/sites-enabled/blog` |
| 阅读统计数据库 | `/home/ubuntu/blog-views.db` |
