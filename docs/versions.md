# 版本基准

本表是全项目唯一的版本事实来源。所有评估结论、ADR、验收记录都以本表版本为准；与本表不一致的结论视为无效。

数据来源：GitHub REST API `GET /repos/{owner}/{repo}/commits/{tag}`，查询时间 2026-11-16（本地时区 America/Los_Angeles）。commit 日期为 commit committer date（UTC）。

| 项目 | 仓库 | 版本 | commit sha | commit 日期 (UTC) |
| --- | --- | --- | --- | --- |
| 原应用 小果影视 | Xanerix/guoapp | `app-v2.0.1-113` | `4c7aab6a87a12e1aaf6b1d08fa31d3844148650c` | 2026-11-16T16:21:19Z |
| 桌面框架 | egoist/mygo | `v0.9.2` | `25ce62f4d357cd6d29da04300e3fbc352d79daa4` | 2026-11-16T12:54:29Z |
| JS 运行时 | MoeForge/moejs | `v0.1.1` | `b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070` | 2026-11-15T17:05:00Z |

## 锁版规则

1. 拉代码按 tag，不按分支：
   - `git clone --branch app-v2.0.1-113 --depth 1 https://github.com/Xanerix/guoapp.git`
   - `go get github.com/egoist/mygo@v0.9.2`
   - `go get github.com/MoeForge/moejs@v0.1.1`
2. `go.mod` 中三者版本固定写死，禁止 `@latest`。
3. 升级任一版本必须单独写 ADR，并在 CHANGELOG 记录影响面。
4. 任何文档引用版本时同时写 tag 与短 sha（7 位），只有 tag 视为不可溯源。

## tag 与 main 的差异

<!-- 待填：若发现 tag 代码与 main 分支存在影响结论的差异，在此登记文件名与差异点，并注明是否改变结论。 -->
