# ADR-0002：UI 路线选 Go + webview（带条件）

- 状态：已接受（条件：R1 的 B 段真实源结果不劣于原 Flutter 应用对照组；不达标则退"嵌 mpv"，按单源 136~150 重比——A 单源 133 更省即选 A，重开本 ADR）
- 日期：2026-11-17
- 决策人：Chief（本群当轮拍板）
- 关联：ADR-0003（播放内核）、ADR-0005（IPC）、ADR-0006（前端选型）、[../evaluation.md](../evaluation.md) §5、[../evaluation-plan.md](../evaluation-plan.md) R0

## 背景

- 原应用 `app-v2.0.1-113`（`4c7aab6`）的 19 个页面、27,822 行 UI 在 Flutter 里；业务核心已有 11,784 行非测试 Go（`native/`，47 个 JSON RPC op），通过一个 C ABI 入口被调用，重写要补的是 UI 壳、平台层和进程模型。
- MyGo `v0.9.2`（`25ce62f`）提供两条 UI 路线：原生 `ui` 包（GPU 直绘、272 个控件构造器）与 webview（WKWebView/WebView2，Typed IPC + 自动生成 TS client）。
- 硬约束：macOS arm64 是产品主目标，但团队当前没有 Mac，WebKit 真机行为、`.app`、签名公证均未验证；Windows 11 是当前唯一可验证平台。
- 原生路线的实测缺口：画中画只支持页面内 `<video>`（`page.go` / `WindowOptions.PictureInPicture`），原生视频无 PiP；播放器周边（弹幕、手势、迷你窗）需自绘。
- webview 路线的实测前提已满足一半：WebView2 `157.0.4341.46` 中 `'MediaSource' in window` = true，H.264/H.265 `isTypeSupported` 均为 true；代理 → MSE 链路通（fMP4 首帧 610 ms / seek 35 ms，TS 693 ms / 3 ms，见 [../spike-stack.md](../spike-stack.md) §6）。

## 备选方案

| 方案 | 做法 | 优点 | 代价 | 风险 |
| --- | --- | --- | --- | --- |
| A 全 Go 原生 UI | 用 `ui` 包重写 19 页与播放器 | 单二进制、内存低、GPU 直绘、无 webview 进程；一套渲染、一套调试 | 播放器交互与弹幕全部自绘；UI 层跟着 MyGo 的 breaking 走 | 0.9.x 控件库仍在快速演化（83 release）；无原生视频 PiP；团队需学新 UI 库 |
| B Go + webview | Go 做核心/窗口/IPC，前端写页面，播放用 `<video>` + hls.js + 代理 | 复用现有 Go 核心与代理；PiP/全屏/弹幕为 web 成熟解法；团队栈匹配；UI 与核心解耦可换壳 | webview 进程内存较高；Windows 无原生 HLS，依赖 hls.js；多一层 IPC | webview 内能否稳定播真实源（R1）；macOS 真机未验证 |
| C 原生壳 + webview 播放页 | 列表/设置用原生 `ui`，播放页嵌网页 | 播放页拿到 PiP，页面性能较好 | 两套渲染、两套调试，边界易腐化 | 单人团队维护两套技术栈容易烂尾 |

## 决定

选 **B（Go + webview）**，v1 只做红果单源、M6 由 M6-lite 模块宿主承接，Windows 先行、macOS 挂账。**本决定不给工期数字**：团队不估算工期（用户 2026-11-17 明令），进度以各阶段 AC 全部通过为出口。

三条可复核的理由：

1. 团队现有项目是 Vue/TS（plus-ui 等），页面产能高于从零学一个每周在变的 Go UI 库。
2. 播放器周边（PiP、全屏、字幕、弹幕、手势）在浏览器里是成熟解法；原生 `ui` 的视频 PiP 只覆盖页面 `<video>`，选 A 等于主动放弃这块。
3. 取流已在 Go 侧解决（`app_stream.go` 本地代理 + 自定义头 + `app_stream_cenc.go` CENC 解密），webview 只需吃代理吐出的明文流，不需要重造播放链路。

**本决定带条件**：R1 的 B 段（真实源 × 30 分钟 × 原 Flutter 应用对照组，在用户个人设备执行）结果若不劣于对照组，B 确认；不达标则退回"嵌 mpv/native module"路径，B 失去复用 webview 播放这一主要优势，而 A 保有单进程、单渲染的结构性优势，届时按 [../evaluation-plan.md](../evaluation-plan.md) R0 **改选 A**。

## 后果

- 正面：UI 与 Go 核心解耦，后续换壳不动核心；macOS 上另有 WKWebView 原生 HLS 可用；减少自绘播放器的工作量。
- 负面（明确承认的代价）：多一个 webview 进程的内存开销；必须维护 IPC 与前端构建产物；Windows 的 HLS 依赖 hls.js（代理需提供 CORS）；HEVC/AV1 在 Windows 仍受系统解码器限制。
- 需要同步修改的文档/模块：ADR-0005（IPC 协议）、ADR-0006（前端框架与资源嵌入）、[../architecture.md](../architecture.md)、[../platform.md](../platform.md)；`docs/README.md` 的索引无需改。

## 验证方式

- **A 段（已在 Windows 完成）**：WebView2 中 MSE 可用，代理 → MSE 通，fMP4 与 TS 均可播。证据：[../spike-stack.md](../spike-stack.md) §6。
- **B 段（待验证，用户个人设备）**：3 个真实源 × 30 分钟 × 对照组，判据为首帧 ≤2 s、seek 恢复 ≤1.5 s、内存 ≤400 MB（私有内存计）且不劣于对照组。
- **macOS 真机（待验证）**：按 [../../tools/spike/MAC-RUN.md](../../tools/spike/MAC-RUN.md) 跑一次，销 `spike-stack.md` §5.2 / §9 的挂账；在此之前 macOS 一律写"未验证"。
