+++
date = '2026-02-28T18:00:00+08:00'
draft = false
title = 'Hugo + PaperMod 配置指南：从零到可用'
tags = ['Hugo', '建站', '方法论']
+++

> 本文基于本站的实际配置，记录 Hugo + PaperMod 主题的核心配置项和常见操作。
> Hugo 版本：0.157.0 | 主题：PaperMod

---

## 1. 项目初始化

```bash
# 创建站点
hugo new site blog
cd blog

# 安装 PaperMod 主题（作为 Git Submodule）
git init
git submodule add https://github.com/adityatelange/hugo-PaperMod.git themes/PaperMod
```

然后编辑根目录下的 `hugo.toml`（Hugo 0.110+ 默认用 toml，旧版用 config.yaml）。

---

## 2. hugo.toml 逐项解读

以下是本站完整配置，每项都有注释：

```toml
# --- 基础信息 ---
baseURL = 'https://sharonzhou.site/'   # 你的域名，末尾带 /
languageCode = 'zh-cn'                  # 语言代码，影响 HTML lang 属性
title = 'My Blog'                       # 站点标题，显示在浏览器标签页
theme = 'PaperMod'                      # 主题名，对应 themes/ 下的目录名

# 启用 Git 信息：自动从 git commit 读取文章的创建/修改时间
enableGitInfo = true

# --- Front Matter 日期策略 ---
[frontmatter]
date = ['date', ':git']     # 文章日期：优先用 front matter 里的 date，没有则取 git 首次提交时间
lastmod = [':git']           # 最后修改时间：直接取 git 最近提交时间

# --- 输出格式 ---
[outputs]
home = ["HTML", "RSS", "JSON"]   # 首页生成 HTML + RSS + JSON（JSON 是搜索功能必需的）

# --- 作者信息 ---
# 注意：必须写成字符串，不能用表格格式
# [params.author] + name = "sharon" 会导致页面显示 map[name:sharon]
[params]
author = "sharon"

# --- 站点参数 ---
defaultTheme = "auto"           # 主题模式：auto（跟随系统）/ light / dark
ShowReadingTime = true           # 显示预计阅读时间
ShowPostNavLinks = true          # 文章底部显示「上一篇/下一篇」导航
ShowBreadCrumbs = true           # 显示面包屑导航（Home » Posts » 文章名）
ShowShareButtons = false         # 不显示分享按钮
ShowCodeCopyButtons = true       # 代码块显示一键复制按钮
comments = true                  # 启用评论（需要配合评论系统组件）

# --- 导航栏 Logo ---
[params.label]
icon = "/logo.png"               # 放在 static/logo.png，显示在导航栏左侧
iconHeight = 30                   # logo 高度（px）

# --- 首页 Welcome 区域 ---
[params.homeInfoParams]
Title = "Welcome"                # 首页大标题
Content = "个人博客"              # 首页副标题/描述

# --- 导航菜单 ---
# weight 决定显示顺序，数字越小越靠前
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

---

## 3. 特殊页面配置

Hugo + PaperMod 有几个特殊页面需要手动创建对应的 Markdown 文件：

### 归档页 `content/archives.md`

```yaml
---
title: "归档"
layout: "archives"
summary: "archives"
---
```

`layout: "archives"` 告诉 Hugo 使用 PaperMod 内置的归档模板，按年月分组展示所有文章。

### 搜索页 `content/search.md`

```yaml
---
title: "搜索"
layout: "search"
placeholder: "输入关键词搜索"
---
```

搜索依赖首页输出 JSON（`outputs` 里的 `"JSON"`），PaperMod 会生成 `index.json` 作为搜索索引。

### 关于页 `content/about.md`

```yaml
---
title: "You are the Magic"
---

这里写关于页内容，支持完整的 Markdown 语法。
```

普通页面，没有特殊 layout，用默认的 single 模板渲染。

---

## 4. 文章 Front Matter

每篇文章的头部元数据，Hugo 支持 TOML（`+++`）和 YAML（`---`）两种格式：

### TOML 格式（本站使用）

```toml
+++
date = '2026-02-28T18:00:00+08:00'
draft = false
title = '文章标题'
tags = ['标签1', '标签2']
+++
```

### YAML 格式

```yaml
---
date: 2026-02-28T18:00:00+08:00
draft: false
title: "文章标题"
tags:
  - 标签1
  - 标签2
---
```

### 关键字段说明

| 字段 | 作用 | 注意事项 |
|------|------|----------|
| `date` | 文章发布日期 | 可省略（如果开了 `enableGitInfo`，会用 git 时间） |
| `draft` | 是否为草稿 | `true` = 不发布，`false` = 发布。**拼写必须正确**，`flase` 会导致构建失败 |
| `title` | 文章标题 | 显示在页面和列表中 |
| `tags` | 标签列表 | 用于分类和标签页 |
| `summary` | 摘要 | 可选，不填则自动截取正文前 70 个字 |
| `cover.image` | 封面图 | 可选，PaperMod 会在文章卡片上显示 |

---

## 5. 图片使用

Hugo 中使用图片有三种方式：

### 方式一：放在 `static/` 目录（全局图片）

适合 Logo、favicon 等全站共用的图片。

```
static/
├── logo.png        # → 引用路径 /logo.png
└── images/
    └── banner.jpg  # → 引用路径 /images/banner.jpg
```

Markdown 中引用：

```markdown
![描述](/images/banner.jpg)
```

### 方式二：文章同级目录（Page Bundle，推荐）

把图片和文章放在同一个文件夹，便于管理：

```
content/posts/my-post/
├── index.md          # 文章内容（注意是 index.md 不是 my-post.md）
├── cover.jpg         # 封面图
└── screenshot.png    # 文章内的图片
```

Markdown 中直接用文件名引用：

```markdown
![截图](screenshot.png)
```

Front Matter 配置封面图：

```toml
[cover]
image = "cover.jpg"
alt = "封面图描述"
```

### 方式三：外部图床

直接用完整 URL：

```markdown
![描述](https://example.com/images/photo.jpg)
```

适合大量图片的场景（对象存储如腾讯云 COS、阿里云 OSS）。

### Logo 配置

Logo 放在 `static/logo.png`，然后在 `hugo.toml` 中配置：

```toml
[params.label]
icon = "/logo.png"      # static/ 下的路径
iconHeight = 30          # 高度，单位 px
```

---

## 6. 自定义样式

PaperMod 提供了 `assets/css/extended/` 目录用于自定义 CSS，不需要修改主题文件：

```
assets/css/extended/
└── custom.css    # 你的自定义样式，自动加载
```

自定义字体需要在 `layouts/partials/extend_head.html` 中引入：

```html
<!-- Google Fonts -->
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=..." rel="stylesheet">
```

PaperMod 的覆盖机制：在 `layouts/partials/` 下放同名文件会覆盖主题模板，常用的覆盖点：

| 文件 | 作用 |
|------|------|
| `extend_head.html` | 在 `<head>` 末尾插入内容（字体、meta 标签） |
| `extend_footer.html` | 在 `<body>` 末尾插入内容（JS 脚本） |
| `comments.html` | 评论组件（giscus / utterances） |

---

## 7. 常用命令

```bash
# 本地预览（包含草稿）
hugo server --buildDrafts

# 本地预览（只看已发布）
hugo server

# 生成静态文件到 public/
hugo

# 新建文章
hugo new content posts/my-post.md
```

---

## 8. 常见坑

### draft 拼写错误
`draft = flase` 会导致 TOML 解析失败，hugo 直接报错。这不是"文章没发布"，而是**整个站点构建失败**。

### author 配置格式
```toml
# 正确 ✓ — 直接写成字符串
[params]
author = "sharon"

# 错误 ✗ — 页面会显示 map[name:sharon]
[params.author]
name = "sharon"
```

PaperMod 的 `author.html` 模板期望 `site.Params.author` 是字符串类型。如果用 `[params.author]` 表格格式，Hugo 会解析为 map 对象，模板直接输出 `map[name:sharon]`。

### 文件名与 URL
文件名 = URL 路径。`content/posts/my-post.md` 的 URL 是 `/posts/my-post/`。文件名拼错了 URL 也会错，改文件名等于改 URL，已有的外链会 404。

### Git Submodule
clone 项目后主题目录为空，需要初始化 submodule：
```bash
git submodule update --init --recursive
```

---

## 9. 目录结构总览

```
blog/
├── assets/css/extended/
│   └── custom.css            # 自定义样式
├── content/
│   ├── posts/                # 文章
│   ├── archives.md           # 归档页（需手动创建）
│   ├── search.md             # 搜索页（需手动创建）
│   └── about.md              # 关于页
├── layouts/partials/         # 覆盖主题模板
├── static/                   # 静态文件（图片、favicon 等）
│   └── logo.png              # 导航栏 Logo
├── themes/PaperMod/          # 主题（Git Submodule）
├── hugo.toml                 # 站点配置
└── public/                   # 生成的静态文件（不要提交到 git）
```
