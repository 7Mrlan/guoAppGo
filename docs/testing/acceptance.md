# 验收标准

状态：M0 已登记；M1 子任务已按 `Eval` 断言改写（2026-11-17），其余待写。每个功能落地时在此登记，未登记验收标准的功能不进入发版回归。

## 写法要求

一条验收标准必须能被两个人独立判定出同样结论：

- 前置条件：数据、网络、源状态
- 操作步骤：可复现的顺序，逐步编号，**每一步是命令或 `Eval` 表达式**
- 期望结果：可观测的具体表现（退出码、`Eval` 返回值、文件是否生成、UIA 属性）
- 反例：至少一条失败路径及其期望表现（无网、源失效、文件损坏）

禁止写法："正常显示""体验流畅""基本可用""性能良好""人工确认""看起来切换了""目测"。

### 两条测试规矩（2026-11-17 定，违反会产生假故障）

1. **`Eval` 不许在 `OnDidFinishLoad` 之类的主线程回调里调用**，必须从 goroutine 调；否则会卡死、25 秒超时、只留一行 `did-finish-load`，看起来像随机失败。
2. **GUI 测试用 `MYGO_E2E=1` 开关**：没有桌面会话时（本机受限会话里 WebView2 返回 `0x8000FFFF`）测试 `t.Skip`，进程退出码 0，**跳过不算失败**；有会话时正常断言。

## M1 子任务（前端工程与设计系统，现实 7 人天）

来源：`evaluation-plan.md` §6.3 的 M1（乐观 4 / 现实 7 / 悲观 12 人天）。三条子任务按依赖顺序推进，前一条的出口条件不满足不进入下一条。

| 子任务 | 范围 | 人天 | 出口条件 |
| --- | --- | ---: | --- |
| M1-1 构建与嵌入 | TS 构建链、`web/dist` 原子产出、`go:embed`、`go tool mygo dev` 开发回路 | 1.5 | AC-M1-1a、AC-M1-1b 通过 |
| M1-2 设计令牌与主题 | 颜色/间距/字号/圆角/动效令牌，暗亮主题、系统跟随、手动覆盖 | 2 | AC-M1-2a、AC-M1-2b 通过 |
| M1-3 组件与路由 | 按钮/输入/卡片/列表/空状态/加载/错误 + hash 路由 + 页面骨架 | 3.5 | AC-M1-3a、AC-M1-3b 通过 |

## 清单

| 编号 | 功能 | 前置条件 | 步骤 | 期望结果 | 反例与期望 | 关联需求/ADR | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| AC-M1-1a | 前端构建可复现并嵌入二进制 | 干净克隆；Node ≥22；esbuild 版本锁定 | 1. `npm ci`；2. `npm run build`；3. `go build ./...`；4. `go test ./web -run TestDistEmbedded`；5. 启动二进制 | 1–4 退出码 0；`web/dist/` 含 index.html 与至少 1 个 js；连续两次构建 dist 的 sha256 相同；启动后 `Eval("document.title")` 返回 `guoAppGo` | 断网时 `npm ci` 失败且不产出 dist：构建脚本先写临时目录、成功后原子替换，失败后 `web/dist` 仍是上一版 | ADR-0002；evaluation-plan §6.3 M1 | 未开始 |
| AC-M1-1b | 干净克隆可编译（.gitignore 回归） | 无 | 1. `git check-ignore -v web/dist/index.html`；2. `git ls-files web/dist`；3. 按 `git ls-files -co --exclude-standard` 复制树后 `go build ./...` | 1 退出码 1；2 列出至少 2 个文件；3 退出码 0 | 临时把规则改成非锚定 `dist/` 后，复制树 `go build` 必须失败并报 `pattern all:dist` | M0 阻塞项；evaluation-plan §6.3 M1 | 未开始 |
| AC-M1-2a | 设计令牌单一来源且取值正确 | tokens.css、`scripts/check-tokens.mjs`、`scripts/contrast.mjs` 存在 | 1. `node scripts/check-tokens.mjs`；2. `node scripts/contrast.mjs`；3. `MYGO_E2E=1 go test ./web -run TestM1Tokens -v` | 1 退出 0、令牌数 24；2 对 12 组前景/背景打印 WCAG 比值，正文 ≥4.5、大字 ≥3.0，退出 0；3 中 `Eval("getComputedStyle(document.documentElement).getPropertyValue('--color-bg').trim()")` 等于 tokens.css 里的值，且 `Eval("document.documentElement.dataset.theme")` 为 `dark` 或 `light` | 把 `--color-bg` 改成 `#ffffff` 后 TestM1Tokens 必须 FAIL（断言不等），不是人工看；无桌面会话时测试 SKIP、退出 0 | M1-2 | 未开始 |
| AC-M1-2b | 主题跟随系统且手动覆盖可持久 | 系统主题可切；`MYGO_E2E=1` | 1. `MYGO_E2E=1 go test ./web -run TestM1Theme -v` | 无覆盖时 `Eval("document.documentElement.dataset.theme === (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')")` 为 true；写 `localStorage.theme='dark'` 并 reload 后 `Eval("document.documentElement.dataset.theme")` 为 `dark`，且 `Eval("getComputedStyle(document.body).backgroundColor")` 等于令牌暗色值 | `localStorage.theme='blue'` 时回落跟随系统，`dataset.theme` 仍是 dark/light 之一，body 背景不是 transparent | M1-2 | 未开始 |
| AC-M1-3a | 通用组件计数、可访问名与禁用态正确 | `/dev/components` 页；`MYGO_E2E=1` | 1. `MYGO_E2E=1 go test ./web -run TestM1Components -v` | `Eval("document.querySelectorAll('[data-ac]').length")` 为 12；`Eval("[...document.querySelectorAll('[data-ac]')].filter(e => !e.getAttribute('aria-label') && !e.textContent.trim()).length")` 为 0；`Eval("document.querySelectorAll('[data-ac][disabled]').length")` 为 1；`Eval("document.querySelectorAll('[data-ac]:not([disabled])').length")` 为 11；Esc 后 `Eval("document.querySelector('[data-role=layer]')")` 为 null | 删掉某个控件的 aria-label 后测试必须 FAIL（计数 1）；弹层打开时焦点在弹层内，关闭后回到触发元素（`Eval("document.activeElement.dataset.ac")` 与触发元素相等） | M1-3 | 未开始 |
| AC-M1-3b | 路由点击、直开与刷新保持 | `MYGO_E2E=1`；hash 路由 | 1. `MYGO_E2E=1 go test ./web -run TestM1Routes -v` | 点击 `Eval("document.querySelector('[data-route=explore]').click()")` 后 `Eval("location.hash")` 为 `#/explore` 且 `Eval("document.querySelector('[aria-current=page]').dataset.route")` 为 `explore`；reload 后 hash 仍为 `#/explore`；设 `location.hash='#/nope'` 后 `Eval("!!document.querySelector('[data-role=not-found]')")` 为 true | 未知路由不白屏：`Eval("document.body.children.length")` 大于 0；空路由回落 `#/`（`location.hash` 为空或 `#/`） | M1-3 | 未开始 |

## 平台标注

每条标准若只在单一平台成立，必须在编号后标 `[mac]` / `[win]`；两平台通用不标。当前所有 M1 条目在 Windows 上执行，macOS 未验证。
