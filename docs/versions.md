# 版本基准

本表是全项目唯一的版本事实来源。所有评估结论、ADR、验收记录都以本表版本为准；与本表不一致的结论视为无效。

数据来源：GitHub REST API `/commits` 与本地 `git`，查询时间 2026-11-17（America/Los_Angeles）。**guoapp 基准为 main 分支提交**；MyGo 与 moejs 仍按 tag 钉死。commit 日期为 commit committer date（UTC）。

| 项目 | 仓库 | 版本 | commit sha | commit 日期 (UTC) |
| --- | --- | --- | --- | --- |
| 原应用 小果影视 | Xanerix/guoapp | `main`（版本 `2.0.2+114`；tag `app-v2.0.1-113` 仅作阅读锚点） | `012de5f6ed67969ca57df03ccca1225018427277` | 2026-11-17T16:54:50Z |
| 桌面框架 | egoist/mygo | `v0.9.2` | `25ce62f4d357cd6d29da04300e3fbc352d79daa4` | 2026-11-16T12:54:29Z |
| JS 运行时 | MoeForge/moejs | `v0.1.1` | `b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070` | 2026-11-15T17:05:00Z |

### guoapp 的基准与阅读锚点

`app-v2.0.1-113`（`4c7aab6a87a12e1aaf6b1d08fa31d3844148650c`）只作阅读锚点：评估早期（第 0 版）的数据在它上面采集。当前基准是 main 的 `012de5f6ed67969ca57df03ccca1225018427277`，两者相距 **34 个提交、107 个文件变更（+4,444/−1,349）**（`git rev-list --count` 与 GitHub compare `ahead_by` 两条路径一致）。上游仍在高频提交，因此"基准之后新增了什么"每次开工都要重新取一次。

统计提交距离前必须确认仓库不是 shallow：对完整 clone 执行 `git fetch --depth 1` 会把历史截断，`git rev-list --count` 会给出离谱的数（本次曾因此把 34 算成 401）。用 `git rev-parse --is-shallow-repository` 检查，必要时 `git fetch --unshallow`。

### mygo 的两个来源

同一个 MyGo 有两条引入路径，版本必须一致：

| 来源 | 声明位置 | 当前版本 | 用途 |
| --- | --- | --- | --- |
| Go 模块 | `go.mod`：`require github.com/egoist/mygo v0.9.2`、`tool github.com/egoist/mygo/cmd/mygo` | `v0.9.2` | 运行时、`go tool mygo` 开发命令、类型绑定 |
| npm 包 | `package.json` devDependencies：`mygo-cli 0.9.2`，实装版本由 `package-lock.json` 锁定 | `0.9.2` | 仅 `mygo.config.ts` 的 `defineConfig` 类型与前端侧工具 |

两者是同一项目的两种分发，不是两个依赖：npm 侧的版本号不带 `v` 前缀。任一升级必须同时改另一侧并写 ADR，否则 `mygo.config.ts` 的类型与 `go tool mygo` 的实际行为会分叉。`node scripts/check-docs.mjs` 会校验两边一致。

## 锁版规则

1. 原应用按 main 最新提交拉取，MyGo / moejs 按 tag：
   - `git clone --filter=blob:none https://github.com/Xanerix/guoapp.git && git -C guoapp checkout 012de5f6ed67969ca57df03ccca1225018427277`
   - `go get github.com/egoist/mygo@v0.9.2`
   - `go get github.com/MoeForge/moejs@v0.1.1`
2. `go.mod` 中三者版本固定写死，禁止 `@latest`。
3. 升级任一版本必须单独写 ADR，并在 CHANGELOG 记录影响面。
4. 引用 guoapp 时**至少要出现一次 40 位完整 sha**，且与本表一致；同时写短号便于阅读不违规，但只写短号、或只写 tag，视为不可溯源。`check-docs.mjs` 会校验本表至少含 3 条完整 sha、且其他文档不出现本表之外的 sha。
5. **新增依赖的登记边界**（决定哪些进本表、哪些不进）：
   - 进本表：同一组件有多个来源需要对齐的（如 mygo 的 Go 模块与 npm 包）、选型由 ADR 决定的、被验收标准或体积/性能预算直接约束的。
   - 不进本表：普通 npm 依赖以 `package-lock.json` 为准，Go 依赖以 `go.mod` + `go.sum` 为准；它们由 lock 保证可复现，逐个抄进本表只会制造第二份会过期的副本。
   - 不论进不进表，新增依赖必须在同一次提交里同时出现：ADR（或已有 ADR 的补充说明）、lock 文件、以及 CHANGELOG 中面向使用者的影响（若用户可感知）。
6. **校验器覆盖范围**（避免虚假安全感）：`scripts/check-docs.mjs` 只校验三个上游 tag/sha、mygo 双来源版本一致、以及文档自身的链接/编号/登记；**其他依赖的漂移不由它兜底**，不能因为它是绿的就把新依赖绕过登记边界。

## M1-2 的依赖判定

结论：**M1-2 不引入新依赖，本表不变**。

- AC-M1-2a 要求 `scripts/contrast.mjs` 打印 12 组 WCAG 对比度，正文 ≥4.5、大字 ≥3.0。WCAG 2.x 的相对亮度与对比度公式约 20 行、只用 sRGB 通道线性化，Node 内置 `fs` 足够，不需要 `culori`、`colorjs.io` 之类。
- AC-M1-2a 还要求 `scripts/check-tokens.mjs` 报"令牌数 24"，同样是解析 `tokens.css` 文本，无需解析库。
- 若实现时仍要引入色差/色彩空间库，按锁版规则第 5 条登记进本表，并在 ADR-0006 补一句选型理由；不允许"先装了再说"。

## 基准切换记录（2026-11-17）

- 原应用基准从 tag `app-v2.0.1-113`（`4c7aab6a87a12e1aaf6b1d08fa31d3844148650c`）切到 main 提交 `012de5f6ed67969ca57df03ccca1225018427277`（版本 `2.0.2+114`）。原因：上游仍在高频提交，tag 之后已有 **34 个提交、107 个文件变更（+4,444/−1,349）**，按 tag 收工会持续漏功能。
- 规模在 `012de5f6` 实测（命令见 `evaluation.md` 附录 A）：Dart 198 文件 49,161 行（`lib` 137 / 38,795；`lib/ui` 81 / 28,778；`lib/ui/player` 19 / 9,185）；Go `native/` 87 文件（非测试 77 / 11,798 行）；页面仍为 19 个 `*_page.dart`。
- 与 tag 时点相比：Dart +6 文件 / +1,701 行；Go 非测试 +14 行；`lib/ui` 文件数由 89 降到 81（目录重组，行数仍增）。
- 从本表提交号起算增量，不看 tag；M2 及以后的清单以本表为起点。
