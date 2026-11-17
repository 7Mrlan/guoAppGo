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
| M1-2 设计令牌与主题 | 颜色/间距/字号/圆角/动效令牌，暗亮主题、系统跟随、手动覆盖 | 2 | AC-M1-2a、AC-M1-2b 通过；AC-M1-2c（基线）可与 M1-3 并行，不阻塞 M1-2 出口 |
| M1-3 组件与路由 | 按钮/输入/卡片/列表/空状态/加载/错误 + hash 路由 + 页面骨架 | 3.5 | AC-M1-3a、AC-M1-3b 通过 |

## 清单

| 编号 | 功能 | 前置条件 | 步骤 | 期望结果 | 反例与期望 | 关联需求/ADR | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| AC-M1-1a | 前端构建可复现并嵌入二进制 | 干净克隆；Node ≥22；Vue/Vite 版本由 `package-lock.json` 锁定（ADR-0006） | 1. `npm ci`；2. `npm run build`；3. `git diff --exit-code web/dist`；4. `go build ./...`；5. `go test ./web -run TestDistEmbedded`；6. 启动二进制 | 1–5 退出码 0；3 为空 diff（产物与源码一致）；`web/dist/` 含 index.html 与至少 1 个 js；连续两次构建 dist 的 sha256 相同；首屏 JS ≤200 KB gzip（构建脚本打印）；启动后 `Eval("document.title")` 返回 `guoAppGo` | 断网时 `npm ci` 失败且不产出 dist：构建脚本先写临时目录、成功后原子替换，失败后 `web/dist` 仍是上一版 | ADR-0002、ADR-0006；evaluation-plan §6.3 M1 | 未开始 |
| AC-M1-1b | 干净克隆可编译（.gitignore 回归） | 无 | 1. `git check-ignore -v web/dist/index.html`；2. `git ls-files web/dist`；3. 按 `git ls-files -co --exclude-standard` 复制树后 `go build ./...` | 1 退出码 1；2 列出至少 2 个文件；3 退出码 0 | 临时把规则改成非锚定 `dist/` 后，复制树 `go build` 必须失败并报 `pattern all:dist` | M0 阻塞项；evaluation-plan §6.3 M1 | 未开始 |
| AC-M1-2a | 设计令牌单一来源且取值正确 | tokens.css、`scripts/check-tokens.mjs`、`scripts/contrast.mjs` 存在 | 1. `node scripts/check-tokens.mjs`；2. `node scripts/contrast.mjs`；3. `MYGO_E2E=1 go test ./web -run TestM1Tokens -v` | 1 退出 0、令牌数 24；2 对 12 组前景/背景打印 WCAG 比值，正文 ≥4.5、大字 ≥3.0，退出 0；3 中 `Eval("getComputedStyle(document.documentElement).getPropertyValue('--color-bg').trim()")` 与 tokens.css 里**当前生效主题**声明的值一致，且 `Eval("document.documentElement.dataset.theme")` 与 `matchMedia('(prefers-color-scheme: dark)').matches` 一致 | 把 `--color-bg` 改成 `#ffffff` 后 TestM1Tokens 必须 FAIL（断言不等），不是人工看；无桌面会话时测试 SKIP、退出 0 | M1-2 | 待复核 |
| AC-M1-2b | 主题跟随系统且手动覆盖可持久 | 系统主题可切；`MYGO_E2E=1` | 1. `MYGO_E2E=1 go test ./web -run TestM1Theme -v` | 无覆盖时 `Eval("document.documentElement.dataset.theme === (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')")` 为 true；写 `localStorage.theme='dark'` 并 reload 后 `Eval("document.documentElement.dataset.theme")` 为 `dark`，且 `Eval("getComputedStyle(document.body).backgroundColor")` 等于令牌暗色值 | `localStorage.theme='blue'` 时回落跟随系统，`dataset.theme` 仍是 dark/light 之一，body 背景不是 transparent | M1-2 | 待复核 |
| AC-M1-2c | 对比度基线不退化（机器判定） | `scripts/contrast-baseline.json` 存在，键为 `theme|fg|bg`，24 条与 `PAIRS` 一一对应 | 1. `node scripts/contrast.mjs`；2. `node scripts/contrast.mjs --update` 后再跑；3. 调暗 `--color-success` 10% 后重跑；4. 临时删一条 `PAIRS` 后重跑；5. 新增一条配对但不更新基线后重跑 | 1 退出 0 且不修改基线文件（`git diff --exit-code scripts/contrast-baseline.json` 为空）；2 退出 0 且基线 24 条不变或新增；3 退出 1 并打印“低于基线：<key> 现值 vs 基线”；4 退出 1 打印“基线残留：<key>”；5 退出 1 打印“新配对未入基线：<key>”，提示先跑 `--update` | 不带参数运行时写基线视为失败；`--update` 是唯一写基线的入口 | M1-2；ADR-0006 的产物校验同理 | 未开始 |
| AC-M1-3a | 通用组件计数、可访问名与禁用态正确 | `/dev/components` 页；`MYGO_E2E=1` | 1. `MYGO_E2E=1 go test ./web -run TestM1Components -v` | `Eval("document.querySelectorAll('[data-ac]').length")` 为 12；`Eval("[...document.querySelectorAll('[data-ac]')].filter(e => !e.getAttribute('aria-label') && !e.textContent.trim()).length")` 为 0；`Eval("document.querySelectorAll('[data-ac][disabled]').length")` 为 1；`Eval("document.querySelectorAll('[data-ac]:not([disabled])').length")` 为 11；Esc 后 `Eval("document.querySelector('[data-role=layer]')")` 为 null | 删掉某个控件的 aria-label 后测试必须 FAIL（计数 1）；弹层打开时焦点在弹层内，关闭后回到触发元素（`Eval("document.activeElement.dataset.ac")` 与触发元素相等） | M1-3 | 未开始 |
| AC-M1-3b | 路由点击、直开与刷新保持 | `MYGO_E2E=1`；hash 路由 | 1. `MYGO_E2E=1 go test ./web -run TestM1Routes -v` | 点击 `Eval("document.querySelector('[data-route=explore]').click()")` 后 `Eval("location.hash")` 为 `#/explore` 且 `Eval("document.querySelector('[aria-current=page]').dataset.route")` 为 `explore`；reload 后 hash 仍为 `#/explore`；设 `location.hash='#/nope'` 后 `Eval("!!document.querySelector('[data-role=not-found]')")` 为 true | 未知路由不白屏：`Eval("document.body.children.length")` 大于 0；空路由回落 `#/`（`location.hash` 为空或 `#/`） | M1-3 | 未开始 |

## 对比度配对范围（AC-M1-2a 的判定依据）

- 12 组 = 正文、次要文字、强调色 × 背景、表面；`--color-border` 当前只用于分隔线，WCAG 1.4.11 约束的是"识别控件所必需的视觉边界"，装饰性分隔线不适用，故不纳入。
- 若将来把 `--color-border` 用于按钮、输入框边界，必须先调到 ≥3:1，并同步改本条 AC 再进测试，不许只改实现。
- 新增配色令牌必须同时补对应配对，且不得拉低**已有配对**的比值：以 `scripts/contrast-baseline.json` 记录的 24 组为基线，容差 0.01（详见 AC-M1-2c）。新配对须同时满足 4.5/3.0，并跑 `node scripts/contrast.mjs --update` 写入基线；只满足对比度而未入基线，按 AC-M1-2c 第 5 步判失败。不设全局地板——把实测最低值 5.18 当门槛，会用现状冒充规格，挡住合规的调色。
- **基线落地前，AC-M1-2c 一律按"未开始"引用**；任何地方不得写成"已约束"，那等于用文字冒充机制。落地时由实现方同步把本条状态改为"待复核"，并在同一条提交里给出三组 FAIL 的实跑记录。
- **实现要求**：`tokens.css` 在 `--color-border` 声明处必须带一条注释，写明"对背景约 1.5:1，仅用于分隔线，不承担控件可辨识职责，故不适用 WCAG 1.4.11 的 3:1"。只写"仅作装饰"不够——半年后有人看到 1.5:1 会当成缺陷；没写这条注释视为未完成。

## 平台标注

每条标准若只在单一平台成立，必须在编号后标 `[mac]` / `[win]`；两平台通用不标。当前所有 M1 条目在 Windows 上执行，macOS 未验证。
