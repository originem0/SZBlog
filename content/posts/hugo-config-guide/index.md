---
date: "2026-02-28T18:00:00+08:00"
draft: false
title: "Hugo + PaperMod 配置指南：从零到可用"
tags: ["Hugo", "建站", "方法论"]
---

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
title = 'Sharon 的博客'                # 站点标题，显示在浏览器标签页
theme = 'PaperMod'                      # 主题名，对应 themes/ 下的目录名

# 启用 Git 信息：自动从 git commit 读取文章的创建/修改时间
enableGitInfo = true
# 构建未来日期的文章（默认 false，会导致未来日期的文章不显示）
buildFuture = true
# 列表页每页显示的文章数量
paginate = 20

# --- Front Matter 日期策略 ---
[frontmatter]
date = ['date', ':git']     # 文章日期：优先用 front matter 里的 date，没有则取 git 首次提交时间
lastmod = [':git']           # 最后修改时间：直接取 git 最近提交时间

# --- 输出格式 ---
[outputs]
home = ["HTML", "RSS", "JSON"]   # 首页生成 HTML + RSS + JSON（JSON 是搜索功能必需的）

# --- 站点参数 ---
# 注意：author 必须写成字符串，不能用 [params.author] 表格格式
# 否则页面会显示 map[name:sharon]
[params]
env = "production"                  # 启用 OpenGraph / Twitter Card 等 SEO meta
description = "个人技术博客，记录 AI、建站与方法论"
author = "Sharon"
defaultTheme = "auto"               # 主题模式：auto（跟随系统）/ light / dark
mainSections = ["posts", "ai-daily", "thoughts"]  # 归档页和首页显示哪些 Section
ShowReadingTime = true               # 显示预计阅读时间
ShowPostNavLinks = true              # 文章底部显示「上一篇/下一篇」导航
ShowBreadCrumbs = true               # 显示面包屑导航（Home » Posts » 文章名）
ShowShareButtons = false             # 不显示分享按钮
ShowCodeCopyButtons = true           # 代码块显示一键复制按钮
ShowToc = true                       # 文章详情页显示目录（Table of Contents）
TocOpen = true                       # TOC 默认展开
comments = true                      # 启用评论（需要配合评论系统组件）
hideAuthor = true                    # 隐藏文章 meta 中的作者名

# --- 导航栏 Logo ---
[params.label]
icon = "/logo.png"               # 放在 static/logo.png，显示在导航栏左侧
iconHeight = 30                   # logo 高度（px）

# --- 首页 Welcome 区域 ---
[params.homeInfoParams]
Title = "Welcome"                     # 首页大标题
Content = "写作，坐在椅子前的艺术"     # 首页副标题/描述

# --- 导航菜单 ---
# weight 决定显示顺序，数字越小越靠前
[[menus.main]]
name = "文章"
url = "/posts/"
weight = 1

[[menus.main]]
name = "AI 日报"
url = "/ai-daily/"
weight = 2

[[menus.main]]
name = "想法"
url = "/thoughts/"
weight = 3

[[menus.main]]
name = "归档"
url = "/archives/"
weight = 4

[[menus.main]]
name = "搜索"
url = "/search/"
weight = 5

[[menus.main]]
name = "关于"
url = "/about/"
weight = 6

# --- 代码高亮 ---
[markup.highlight]
noClasses = false       # 输出 CSS class 而非内联 style，便于自定义颜色

# --- Goldmark 渲染器 ---
[markup.goldmark.renderer]
unsafe = true           # 允许 Markdown 中的 raw HTML
```

---

## 3. 新建内容板块（Section）

Hugo 用 `content/` 下的子目录来组织不同板块，每个子目录就是一个 Section，有独立的列表页。

### 创建步骤

1. 新建目录和 Section 首页：

```
content/ai-daily/
└── _index.md    # Section 首页（必须有）
```

`_index.md` 内容：

```yaml
---
title: "AI 日报"
description: "每日 AI 领域动态汇总"
---
```

2. 在 `hugo.toml` 中添加导航菜单项：

```toml
[[menus.main]]
  name = "AI 日报"
  url = "/ai-daily/"
  weight = 2
```

3. 往目录里放 `.md` 文章即可，格式和普通文章一样。

4. 如果需要该板块的文章出现在归档页和首页，把目录名加入 `mainSections`：

```toml
[params]
mainSections = ["posts", "ai-daily", "thoughts"]
```

不配置 `mainSections` 时，Hugo 会自动选择文章数最多的 Section，可能导致归档页只显示部分板块。

### 多板块示例

```
content/
├── posts/          # 博客文章 → /posts/
├── ai-daily/       # AI 日报   → /ai-daily/
├── thoughts/       # 短想法    → /thoughts/
├── notes/          # 随笔      → /notes/
└── projects/       # 项目      → /projects/
```

每个 Section 的 URL 就是目录名，PaperMod 会自动为每个 Section 生成列表页。

---

## 4. 特殊页面配置

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

## 5. 文章 Front Matter

每篇文章的头部元数据，Hugo 支持 YAML（`---`）和 TOML（`+++`）两种格式：

### YAML 格式（本站使用）

```yaml
---
date: "2026-02-28T18:00:00+08:00"
draft: false
title: "文章标题"
tags: ["标签1", "标签2"]
---
```

### TOML 格式

```toml
+++
date = '2026-02-28T18:00:00+08:00'
draft = false
title = '文章标题'
tags = ['标签1', '标签2']
+++
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

## 6. 图片使用

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

```yaml
---
cover:
  image: "cover.jpg"
  alt: "封面图描述"
---
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

## 7. 自定义样式

PaperMod 提供了 `assets/css/extended/` 目录用于自定义 CSS，不需要修改主题文件。PaperMod 会自动加载该目录下所有 CSS 文件（按文件名字母序）：

```
assets/css/extended/
├── 01-variables.css    # 亮色/暗色模式 CSS 变量
├── 02-typography.css   # 字体栈、排版、标题大小
├── 03-components.css   # 链接、blockquote、卡片、标签、导航等组件
├── 04-code.css         # chroma 语法高亮（亮色/暗色）
├── 05-sidebar.css      # 侧边栏导航 + TOC 浮动
└── 06-responsive.css   # 响应式 media queries
```

自定义字体需要在 `layouts/partials/extend_head.html` 中引入（本站使用 Noto Sans SC + Inter + JetBrains Mono）：

```html
<!-- Google Fonts: Noto Sans SC (中文), Inter (Latin), JetBrains Mono (代码) -->
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600&family=JetBrains+Mono:wght@400&family=Noto+Sans+SC:wght@400;600&display=swap" rel="stylesheet">
```

PaperMod 的覆盖机制：在 `layouts/partials/` 下放同名文件会覆盖主题模板，常用的覆盖点：

| 文件 | 作用 |
|------|------|
| `extend_head.html` | 在 `<head>` 末尾插入内容（字体、meta 标签） |
| `extend_footer.html` | 在 `<body>` 末尾插入内容（JS 脚本） |
| `comments.html` | 评论组件（giscus / utterances） |

### PaperMod CSS 变量覆盖（GitHub 配色）

在 `01-variables.css` 的 `:root` 中覆盖 PaperMod 原生变量，可以全站生效。本站使用 GitHub 配色：

```css
:root {
    --primary: #1f2328;                    /* 正文字色 */
    --secondary: #656d76;                  /* 副文本、时间、摘要 */
    --theme: #ffffff;                      /* 页面背景（纯白） */
    --entry: #ffffff;                      /* 卡片背景 */
    --border: #d0d7de;                     /* 边框 */
    --code-bg: rgba(175, 184, 193, 0.2);   /* 行内代码背景 */
    --code-block-bg: #f6f8fa;              /* 代码块背景 */
}
```

暗色模式用 `[data-theme="dark"]` 选择器单独覆盖（GitHub Dark 配色）：

```css
:root[data-theme="dark"] {
    --primary: #e6edf3;
    --secondary: #8b949e;
    --theme: #0d1117;
    --entry: #161b22;
    --border: #30363d;
    --code-bg: rgba(110, 118, 129, 0.4);
    --code-block-bg: #161b22;
}
```

### 文章目录（TOC）

在 `hugo.toml` 中启用：

```toml
[params]
ShowToc = true       # 显示目录
TocOpen = true       # 默认展开
```

PaperMod 默认在文章顶部显示目录。本站通过 CSS 将 TOC 改为宽屏右侧浮动（>= 1280px），窄屏保持顶部内联。

### 代码高亮主题

Hugo 默认用内联样式渲染代码高亮，要用 CSS 控制必须关闭内联样式：

```toml
[markup.highlight]
noClasses = false    # 输出 CSS class 而非内联 style

[markup.goldmark.renderer]
unsafe = true        # 允许 Markdown 中的 raw HTML（日报等需要内嵌 HTML 的页面）
```

然后在 `assets/css/extended/` 下的 CSS 文件中用 `.chroma` 相关类名控制语法颜色。可以用 Hugo 命令生成主题：

```bash
# 生成 GitHub 亮色主题 CSS
hugo gen chromastyles --style=github > syntax-light.css

# 可选主题：monokai, dracula, github, solarized-light 等
hugo gen chromastyles --style=monokai > syntax-dark.css
```

---

## 8. 常用命令

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

## 9. 常见坑

### buildFuture 默认关闭
Hugo 默认 `buildFuture = false`，如果文章的 `date` 是未来时间（包括时区换算后），构建时会被静默跳过。表现：文章已推送但网站不显示，也不报错。解决：在 `hugo.toml` 中加 `buildFuture = true`。

### draft 拼写错误
`draft: flase` 会导致 YAML 解析失败，hugo 直接报错。这不是"文章没发布"，而是**整个站点构建失败**。

### author 配置格式
```toml
# 正确 ✓ — 直接写成字符串
[params]
author = "Sharon"

# 错误 ✗ — 页面会显示 map[name:Sharon]
[params.author]
name = "Sharon"
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

## 10. 目录结构总览

```
blog/
├── assets/css/extended/       # 自定义样式（PaperMod 按文件名字母序自动加载）
│   ├── 01-variables.css       # 亮色/暗色模式 CSS 变量（GitHub 配色）
│   ├── 02-typography.css      # 字体栈、排版、标题大小
│   ├── 03-components.css      # 链接、blockquote、表格、标签、导航等
│   ├── 04-code.css            # chroma 语法高亮（亮色 GitHub / 暗色 Catppuccin）
│   ├── 05-sidebar.css         # 侧边栏导航 + TOC 浮动
│   └── 06-responsive.css      # 响应式 media queries
├── content/
│   ├── posts/                 # 博客文章
│   ├── ai-daily/              # AI 日报板块
│   │   └── _index.md          # Section 首页（必须有）
│   ├── thoughts/              # 短想法
│   │   └── _index.md          # Section 首页
│   ├── archives.md            # 归档页（需手动创建）
│   ├── search.md              # 搜索页（需手动创建）
│   └── about.md               # 关于页
├── layouts/
│   ├── _default/
│   │   └── archives.html      # 自定义归档模板（隐藏年份和作者）
│   ├── partials/              # 覆盖主题模板
│   │   ├── extend_head.html   # <head> 末尾（字体、meta）
│   │   ├── extend_footer.html # <body> 末尾（JS、侧边栏）
│   │   └── comments.html      # 评论组件
├── static/                    # 静态文件（图片、favicon 等）
│   ├── logo.png               # 导航栏 Logo
│   ├── favicon.ico            # 浏览器标签页图标
│   ├── favicon-16x16.png
│   ├── favicon-32x32.png
│   └── apple-touch-icon.png
├── webhook/                   # Webhook 服务（Go，部署在服务器）
│   └── main.go                # GitHub push → 自动构建
├── themes/PaperMod/           # 主题（Git Submodule，不要修改）
├── hugo.toml                  # 站点配置
└── public/                    # 生成的静态文件（不要提交到 git）
```
