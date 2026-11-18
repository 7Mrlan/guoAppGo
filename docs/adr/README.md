# ADR 索引

一个技术决定一份 ADR。文件命名：`docs/adr/NNNN-短横线小写标题.md`，编号四位递增，不复用、不重排。

状态取值：`提议` / `已接受`（可带条件）/ `已废弃` / `被 ADR-NNNN 取代` / `不适用`（仅限明确作废的条目）。判定细则见文末维护规则。

模板：`docs/adr/0000-template.md`。

## 标题清单（第一批）

状态见下表；`ADR-0002`、`ADR-0003` 已于 2026-11-17 由 Chief 定为 `已接受`（0002 带条件，条件同时写在正文状态行与下表），正文已落。

| 编号 | 标题 | 待回答的问题 | 依赖的输入 | 状态 |
| --- | --- | --- | --- | --- |
| ADR-0001 | 桌面框架与运行时选型：MyGo + moejs 是否采用 | 两者版本、许可、维护活跃度是否满足长期维护；能否只用一个 | evaluation.md §2、§3；spike-stack.md | 提议 |
| ADR-0002 | UI 路线：全 Go 原生 UI / Go + webview / Go + moejs 混合 | 三条路线各自的能力缺口与技术债 | evaluation.md §5、evaluation-plan.md R0；A 段实测 | 已接受（条件：R1 的 B 段真实源不劣于对照组；不达标则退嵌 mpv，B 失去复用 webview 播放这一主要优势，改选 A）：正文 [`0002-ui-route-webview.md`](0002-ui-route-webview.md) |
| ADR-0003 | 视频播放内核 | 用系统媒体框架、自带解码还是 webview `<video>` + hls.js | spike-stack.md §5、§6 | 已接受（R1 不达标则重开本 ADR 改选嵌 mpv）：正文 [`0003-playback-kernel-hlsjs-proxy.md`](0003-playback-kernel-hlsjs-proxy.md) |
| ADR-0004 | 流媒体协议与格式支持范围 | HLS/m3u8、DASH、直链 mp4、字幕轨，各自支持到什么程度 | feature-inventory.md 中播放相关条目 | 提议 |
| ADR-0005 | 主进程与前端/JS 的 IPC 协议 | 消息格式、版本字段、请求响应、事件、超时与错误码 | ADR-0002 已选 webview，本条为 v1 必做 | 提议 |
| ADR-0006 | 前端技术选型与资源打包 | 框架、构建产物、静态资源如何嵌入单文件 | ADR-0002 | **已接受**：Vue 3 + TS + Vite，产物进库；正文 [`0006-frontend-vue-vite.md`](0006-frontend-vue-vite.md) |
| ADR-0007 | 数据持久化：存储形态与迁移 | SQLite 还是文件；schema 版本与迁移策略；写入时机与损坏恢复 | feature-inventory.md 中收藏/历史/设置 | 提议 |
| ADR-0008 | 网络层与源解析 | HTTP 客户端、并发上限、超时/重试、UA/Referer、代理、限流 | evaluation.md 中原应用解析逻辑 | 提议 |
| ADR-0009 | 是否保留插件/脚本能力 | moejs 到底承载什么；不要它的代价；保留它的沙箱与权限边界 | evaluation.md §3 | **不适用**：v1 单源，模块宿主按 M6-lite 只复用现有 goja、不引入第二引擎；v2 恢复多源时另立 |
| ADR-0010 | macOS 分发：签名、公证、打包 | Developer ID、notarytool、DMG、首次启动拦截的处理 | platform.md P-01/P-02 | 提议（无 Mac，暂挂账） |
| ADR-0011 | Windows 分发：签名与安装包 | 是否买代码签名证书；安装包格式；SmartScreen 影响 | platform.md P-01 | 提议 |
| ADR-0012 | 自动更新机制 | 更新源、版本比较、增量与全量、失败回滚 | platform.md P-03 | 提议 |
| ADR-0013 | 平台差异适配层形态 | 差异集中在适配层还是各模块内部分支 | platform.md 全表 | 提议 |
| ADR-0014 | 日志与崩溃诊断 | 落盘位置、轮转、隐私字段、诊断包导出方式 | platform.md P-08 | 提议 |
| ADR-0015 | 测试策略与质量门 | go test 覆盖率口径、哪些包必须有测试、验收标准格式、发版回归范围 | testing/ | 提议 |
| ADR-0016 | 配置与状态管理 | 配置模型、默认值、导入导出、与持久化的边界 | feature-inventory.md 设置块 | 提议 |

## 维护规则

- 决定变更时新建 ADR 并将旧的标为"被取代"，不原地改写历史决定。
- ADR 正文出现"计划/预计"即视为未完成，改写为决定或标 **待验证**。
- 引用依赖版本必须写 tag + 短 sha（见 [../versions.md](../versions.md)），写"最新版"视为未完成。
- 状态取值统一用：`提议` / `已接受` / `已废弃` / `被 ADR-NNNN 取代`；带前提的决定写成"已接受（条件：…）"，条件写进正文而非只写在索引。
- `不适用` 只用于明确作废的条目，必须在正文或索引里写清作废条件（如 ADR-0009 挂在"v1 单源"上）。
