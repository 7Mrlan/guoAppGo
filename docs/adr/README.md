# ADR 索引

一个技术决定一份 ADR。文件命名：`docs/adr/NNNN-短横线小写标题.md`，编号四位递增，不复用、不重排。

状态取值：`提议` / `已接受` / `已废弃` / `被 ADR-NNNN 取代`。

模板：`docs/adr/0000-template.md`。

## 标题清单（第一批）

状态均为 `提议`，待评估结论与用户确认后逐条落正文。

| 编号 | 标题 | 待回答的问题 | 依赖的输入 |
| --- | --- | --- | --- |
| ADR-0001 | 桌面框架与运行时选型：MyGo + moejs 是否采用 | 两者版本、许可、维护活跃度是否满足长期维护；能否只用一个 | evaluation.md §2、§3；spike-stack.md |
| ADR-0002 | UI 路线：全 Go 原生 UI / Go + webview / Go + moejs 混合 | 三条路线各自的能力缺口与工时差 | evaluation.md §4、§5；feature-inventory.md |
| ADR-0003 | 视频播放内核 | 用系统媒体框架（AVFoundation / Media Foundation）还是自带解码（ffmpeg）还是 webview `<video>` | spike 验证结果；需支持的编码/容器清单 |
| ADR-0004 | 流媒体协议与格式支持范围 | HLS/m3u8、DASH、直链 mp4、字幕轨，各自支持到什么程度 | feature-inventory.md 中播放相关条目 |
| ADR-0005 | 主进程与前端/JS 的 IPC 协议 | 消息格式、版本字段、请求响应、事件、超时与错误码 | 仅当 ADR-0002 选 webview 或混合时需要 |
| ADR-0006 | 前端技术选型与资源打包（若走 webview） | 框架、构建产物、静态资源如何嵌入单文件 | ADR-0002 |
| ADR-0007 | 数据持久化：存储形态与迁移 | SQLite 还是文件；schema 版本与迁移策略；写入时机与损坏恢复 | feature-inventory.md 中收藏/历史/设置 |
| ADR-0008 | 网络层与源解析 | HTTP 客户端、并发上限、超时/重试、UA/Referer、代理、限流 | evaluation.md 中原应用解析逻辑 |
| ADR-0009 | 是否保留插件/脚本能力 | moejs 到底承载什么；不要它的代价；保留它的沙箱与权限边界 | evaluation.md §3 |
| ADR-0010 | macOS 分发：签名、公证、打包 | Developer ID、notarytool、DMG、首次启动拦截的处理 | platform.md P-01/P-02 |
| ADR-0011 | Windows 分发：签名与安装包 | 是否买代码签名证书；安装包格式；SmartScreen 影响 | platform.md P-01 |
| ADR-0012 | 自动更新机制 | 更新源、版本比较、增量与全量、失败回滚 | platform.md P-03 |
| ADR-0013 | 平台差异适配层形态 | 差异集中在适配层还是各模块内部分支 | platform.md 全表 |
| ADR-0014 | 日志与崩溃诊断 | 落盘位置、轮转、隐私字段、诊断包导出方式 | platform.md P-08 |
| ADR-0015 | 测试策略与质量门 | go test 覆盖率口径、哪些包必须有测试、验收标准格式、发版回归范围 | testing/ |
| ADR-0016 | 配置与状态管理 | 配置模型、默认值、导入导出、与持久化的边界 | feature-inventory.md 设置块 |

## 维护规则

- 决定变更时新建 ADR 并将旧的标为"被取代"，不原地改写历史决定。
- ADR 正文出现"计划/预计"即视为未完成，改写为决定或标 **待验证**。
- 引用依赖版本必须写 tag + 短 sha（见 [../versions.md](../versions.md)），写"最新版"视为未完成。
