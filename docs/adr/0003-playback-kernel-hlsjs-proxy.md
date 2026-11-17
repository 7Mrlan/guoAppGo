# ADR-0003：播放内核用 hls.js + Go 代理，暂不嵌 mpv

- 状态：已接受（条件：随 ADR-0002 的 R1 B 段；真实源不达标则重开本 ADR，改选内嵌 mpv，预算 +20~30 人天）
- 日期：2026-11-17
- 决策人：Chief（本群当轮拍板）
- 关联：ADR-0002（UI 路线）、ADR-0004（协议范围）、[../evaluation-plan.md](../evaluation-plan.md) R0/R1

## 背景

- 原应用的播放器不在 Go 里：`media_kit` 1.2.6（libmpv）+ `video_player`，解码器选择、硬解/软解、轨切换在 `lib/playback_session.dart`（1,397 行编排）。这是全项目最大的单点不确定。
- 站源需要自定义 `Referer`/`User-Agent`，部分源走 CENC 解密；这两件事现有 Go 代码已做：`app_stream.go`（本地 HTTP 代理 + 自定义头）、`app_stream_cenc.go`（密钥协商与解密）。
- MyGo 原生 `ui/video` 走系统媒体框架（AVFoundation / Media Foundation），**`video.Open/OpenURL/OpenData` 没有请求头参数**，也拿不到 CENC 明文流，直接用它就得在 Go 里再实现鉴权与解密后的本地服务——等于绕一圈回到代理。
- MyGo webview 能开 `<video>`，Windows 侧 MSE 实测可用；Safari/WKWebView 的 MSE 不吃 MPEG-TS，但 hls.js 会在 JS 里把 TS 转封装成 fMP4 再交 MSE（实测 TS 首帧 693 ms、seek 3 ms）。

## 备选方案

| 方案 | 做法 | 优点 | 代价 | 风险 |
| --- | --- | --- | --- | --- |
| A 系统媒体框架直连 | `ui/video` + AVFoundation/Media Foundation | 原生解码、硬件加速、无 JS | 无请求头、无 CENC、无 PiP；需另做鉴权与解密后的本地服务 | 基本要重造现有代理，收益低 |
| B hls.js + Go 代理 | CENC/鉴权在 Go 代理完成，代理吐明文 HLS；webview 用 hls.js + `<video>` | 复用现有代理；PiP/全屏/字幕/弹幕成熟；TS 由 hls.js 转封装 | MSE 内存；依赖系统解码器；hls.js 体积 | B 段真实源表现未验；macOS 真机未验 |
| C 内嵌 mpv/ffmpeg | native module 嵌 libmpv 或 ffmpeg | 编解码覆盖最全，能兜底异常封装 | +20~30 人天；引入 cgo/本地库；GPL 与签名、体积问题 | 与 MyGo"纯 Go、无 cgo"的设计冲突，维护成本高 |

## 决定

选 **B**：Go 代理负责鉴权、自定义头与 CENC 解密，输出明文 HLS；webview 用 **hls.js + `<video>`** 播放。**v1 暂不嵌 mpv**，也不做手选解码器/软硬解切换，交给系统。

两条可复核的理由：

1. 现有代理已经解决了"取流"里最脏的部分（头、签名、CENC），webview 只需要一个干净的 HLS 源。
2. 容器格式不构成额外成本：TS 由 hls.js 在 JS 内转封装（实测首帧 693 ms、seek 3 ms），"代理转封装 +5~8 人天"只在放弃 hls.js、改走原生 `<video>` HLS 时才付。

**退路**：R1 的 B 段真实源若不达标（首帧、seek、内存或稳定性任一不过且对照组正常），重开本 ADR 改选 C，预算按 +20~30 人天、B 变 126~136 处理。

## 后果

- 正面：复用 Go 代理；播放器周边能力（PiP、全屏、字幕、弹幕、手势）用 web 方案；前端与播放逻辑同栈，迭代快。
- 负面（明确承认的代价）：HEVC/AV1 在 Windows 依赖系统解码器（可能需要用户装扩展）；MSE 有额外内存；不支持 DRM（Widevine/PlayReady），v1 不涉及，若将来要支持需重开本 ADR；hls.js 需随包分发（Apache-2.0，注意保留许可声明）。
- 需要同步修改的文档/模块：[../platform.md](../platform.md) 的硬解/编码项、[../architecture.md](../architecture.md) 的播放数据流、ADR-0004（协议范围）、ADR-0006（资源嵌入，含 hls.js）。

## 验证方式

- **容器与链路（Windows，已完成）**：fMP4 首帧 610 ms、seek 35 ms；TS 693 ms、3 ms，均正常播放。证据：[../spike-stack.md](../spike-stack.md) §6。注意 610 ms 来自刻意造的最坏代理（8 KiB 块 + 200 µs/块节流，直连 152~371 ms），真实代理须按大块、零节流、内存直通设计。
- **真实源（待验证，用户个人设备）**：B 段 3 个源 × 30 分钟 × 原 Flutter 应用对照组，判据见 R0/R1。
- **macOS 真机（待验证）**：按 [../../tools/spike/MAC-RUN.md](../../tools/spike/MAC-RUN.md) 跑，重点看 `'MediaSource' in window` 与首帧/seek；在此之前 macOS 播放行为一律写"未验证"。
