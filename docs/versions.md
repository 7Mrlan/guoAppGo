# 版本基准

本表是全项目唯一的版本事实来源。所有评估结论、ADR、验收记录都以本表版本为准；与本表不一致的结论视为无效。

数据来源：GitHub REST API `GET /repos/{owner}/{repo}/commits/{tag}`，查询时间 2026-11-16（本地时区 America/Los_Angeles）。commit 日期为 commit committer date（UTC）。

| 项目 | 仓库 | 版本 | commit sha | commit 日期 (UTC) |
| --- | --- | --- | --- | --- |
| 原应用 小果影视 | Xanerix/guoapp | `app-v2.0.1-113` | `4c7aab6a87a12e1aaf6b1d08fa31d3844148650c` | 2026-11-16T16:21:19Z |
| 桌面框架 | egoist/mygo | `v0.9.2` | `25ce62f4d357cd6d29da04300e3fbc352d79daa4` | 2026-11-16T12:54:29Z |
| JS 运行时 | MoeForge/moejs | `v0.1.1` | `b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070` | 2026-11-15T17:05:00Z |

### mygo 的两个来源

同一个 MyGo 有两条引入路径，版本必须一致：

| 来源 | 声明位置 | 当前版本 | 用途 |
| --- | --- | --- | --- |
| Go 模块 | `go.mod`：`require github.com/egoist/mygo v0.9.2`、`tool github.com/egoist/mygo/cmd/mygo` | `v0.9.2` | 运行时、`go tool mygo` 开发命令、类型绑定 |
| npm 包 | `package.json` devDependencies：`mygo-cli 0.9.2`，实装版本由 `package-lock.json` 锁定 | `0.9.2` | 仅 `mygo.config.ts` 的 `defineConfig` 类型与前端侧工具 |

两者是同一项目的两种分发，不是两个依赖：npm 侧的版本号不带 `v` 前缀。任一升级必须同时改另一侧并写 ADR，否则 `mygo.config.ts` 的类型与 `go tool mygo` 的实际行为会分叉。`node scripts/check-docs.mjs` 会校验两边一致。

## 锁版规则

1. 拉代码按 tag，不按分支：
   - `git clone --branch app-v2.0.1-113 --depth 1 https://github.com/Xanerix/guoapp.git`
   - `go get github.com/egoist/mygo@v0.9.2`
   - `go get github.com/MoeForge/moejs@v0.1.1`
2. `go.mod` 中三者版本固定写死，禁止 `@latest`。
3. 升级任一版本必须单独写 ADR，并在 CHANGELOG 记录影响面。
4. 任何文档引用版本时同时写 tag 与短 sha（7 位），只有 tag 视为不可溯源。
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

## tag 与 main 的差异

<!-- 待填：若发现 tag 代码与 main 分支存在影响结论的差异，在此登记文件名与差异点，并注明是否改变结论。 -->
