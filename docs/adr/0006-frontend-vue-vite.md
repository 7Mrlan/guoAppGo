# ADR-0006：前端用 Vue 3 + TypeScript + Vite，产物进库并由 CI 校验

- 状态：已接受（2026-11-17，随 M1-1 开工一并复核；不改变 ADR-0002 的路线）
- 日期：2026-11-17
- 决策人：探路（技术方案），Chief 授权 M1 开工
- 关联：ADR-0002（UI 路线）、ADR-0005（IPC）、`docs/testing/acceptance.md` AC-M1-1a/1b

## 背景

- ADR-0002 选了 Go + webview，19 个页面与播放器周边都要在前端实现；团队现有项目是 Vue/TS（plus-ui 等），不是从零学框架。
- M0 骨架的约束是"不依赖 Bun、不依赖 Vite"——那是最小骨架的取舍，不是长期选型；M1 开始要正经的构建链。
- 二进制用 `//go:embed all:dist` 打包前端（`web/web.go`），所以构建产物必须能被 `go build` 直接找到；验收要求干净克隆不装 Node 也能编（AC-M1-1b）。
- 约束：Windows 11 为当前唯一可验证平台；macOS 未验证；单二进制分发，前端资源不能依赖外部 CDN。

## 备选方案

| 方案 | 做法 | 优点 | 代价 | 风险 |
| --- | --- | --- | --- | --- |
| A Vue 3 + TS + Vite | SFC 写页面，Vite 构建到 `web/dist`，`base: './'`，hash 路由 | 团队栈匹配；生态成熟（hls.js、vueuse）；构建配置少 | 引入 Node 构建链；产物需进库或 CI 先构建 | Vue 运行时体积；需约束 chunk 预算 |
| B 纯 TS + esbuild | 无框架，手写路由与组件 | 产物最小、构建最快 | 19 个页面与状态管理全部手写；团队产能低 | 后期重构成本高 |
| C Svelte + Vite | 编译型框架，运行时小 | 产物小、写法简洁 | 团队不熟；招人与维护风险 | 学习曲线换来的体积收益有限 |

## 决定

选 **A：Vue 3 + TypeScript + Vite**。三条可复核的理由：

1. 团队已有 Vue/TS 项目，页面产能是 B 路线的主要收益来源；换框架等于自废 ADR-0002 的第一条理由。
2. `base: './'` + hash 路由可以同时适配 `file://` 与本地 HTTP（webview 两种加载方式都不用改代码）。
3. hls.js、vueuse 等依赖在 Vue 生态里是一等公民，播放器周边不需要自己造轮子。

打包规则（与验收标准绑定）：

- `vite.config.ts` 设 `base: './'`，产物输出到 `web/dist`；`web/dist` **进库**，保证干净克隆直接 `go build`。
- `package-lock.json` 进库，`npm ci` 为唯一安装方式；不使用 Bun。
- CI 校验产物与源码一致：`npm ci && npm run build && git diff --exit-code web/dist`，不一致即失败。
- 体积预算：首屏 JS ≤ 200 KB gzip；`hls.js` 独立 chunk，仅在播放页动态 import。
- 开发回路：`npm run dev`（Vite 5173）+ `go tool mygo dev`，仅开发环境指向 dev server；发布产物走 `web/dist`。

## 后果

- 正面：页面开发回到团队熟悉的技术栈；构建产物可复现、可审查；播放器依赖按需加载。
- 负面（明确承认的代价）：仓库里同时存在源码与产物，改前端必须重新构建并一起提交；Node 版本成为构建前置（不进入运行时）；Vue 运行时约 40 KB gzip 计入首屏预算。
- 需要同步修改的文档/模块：`docs/architecture.md` 的前端分层、ADR-0005（IPC client 生成方式）、`docs/testing/acceptance.md` 的 AC-M1-1a/1b、`CHANGELOG.md` 记录规则。

## 验证方式

- **AC-M1-1a**：`npm ci` → `npm run build` → `go build ./...` 退出码 0；连续两次构建 `web/dist` 的 sha256 相同；`go test ./web -run TestDistEmbedded` 通过。
- **AC-M1-1b**：干净克隆（不装 Node）`go build ./...` 退出码 0；`git check-ignore -v web/dist/index.html` 退出码 1。
- **一致性**：CI 上 `npm run build && git diff --exit-code web/dist` 通过；手工改 dist 不重建会被这条拦住。
- **macOS**：未验证；随 B 段的 macOS 挂账一起处理，不把 Windows 的构建结论外推。
