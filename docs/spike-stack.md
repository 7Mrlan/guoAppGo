# 可行性验证（spike）结果

## 0. 版本基准

| 项目 | tag | commit | 拉取方式 |
|---|---|---|---|
| egoist/mygo | `v0.9.2` | `25ce62f` | `git clone --branch v0.9.2 --depth 1` |
| MoeForge/moejs | `v0.1.1` | `b7f49c1` | `git clone --branch v0.1.1 --depth 1` |
| Go 工具链 | `go1.27.4` | — | 官方 zip，sha256 `178ab989…de1e7e` 校验通过 |

验证机器：公司 Windows 11（`10.0.22631`，x64）。**本文件所有 macOS 结论都来自 WebKit 源码，没有在 Mac 上实跑**，见 §5.2 与 §9。

版本口径以 [versions.md](versions.md) 为准；探针源码与其跑法见 [tools/spike/README.md](../tools/spike/README.md)。

## 1. 环境阻塞与绕过方式（可复现）

这台机器上每一步都有坑，先把绕过方式固定下来，否则换个会话谁都得再踩一遍：

1. **系统 HTTPS 出不去**：`curl`、`Invoke-WebRequest` 全部失败。
   `curl -x http://127.0.0.1:7890 https://github.com` 报 `schannel: AcquireCredentialsHandle failed: SEC_E_NO_CREDENTIALS (0x8009030E)`；
   `Invoke-WebRequest` 报"基础连接已经关闭"。可用通道是 node 的 TLS（`NODE_USE_ENV_PROXY=1` + `HTTPS_PROXY=http://127.0.0.1:7890`）与 `git -c http.proxy=…`。
2. **go 的默认目录不可写**：`%TEMP%` 与 `%USERPROFILE%\go` 被拒（`mkdir …Temp\go-build…: Access is denied`）。
   必须显式指定 `GOTMPDIR`、`GOCACHE`、`GOPATH` 到工作区内目录。只改 `GOMODCACHE` 不改 `GOPATH` 时，go 还会去写 `%USERPROFILE%\go\pkg\sumdb\sum.golang.org\latest` 并失败。
3. **GUI 进程启动被拒**：`Start-Process`、cua 的 `launch_app` 返回 `0x800704C7`（ERROR_CANCELLED）。
   可用方式：`Start-Job { & .\app.exe }`，或 `Invoke-CimMethod Win32_Process Create`。
4. **WebView2 在受限会话里建不起来**：前台直跑时日志是
   `mygo: creating the WebView2 controller failed (HRESULT 0x8000ffff); trying again`，三次后失败。
   换成脱离该会话启动（`Win32_Process Create`）后一次成功。这一条与 mygo 无关，CI／服务化环境会再遇到。

## 2. MyGo v0.9.2 (25ce62f)：能起窗口、能跑测试

```powershell
git -c http.proxy=http://127.0.0.1:7890 clone --branch v0.9.2 --depth 1 https://github.com/egoist/mygo
cd mygo; go build -o ..\bin\mygo.exe ./cmd/mygo      # exit 0
mygo init -template native probe-native              # exit 0
```

生成结果：

- `go.mod` 内含 `tool github.com/egoist/mygo/cmd/mygo` 与 `require github.com/egoist/mygo v0.9.2`。
  **模板写的是 `cmd/mygo@latest`**，本次恰好解析到 v0.9.2；要钉死版本得自己 `go get -tool github.com/egoist/mygo/cmd/mygo@v0.9.2`。
- `mygo.json`：`name` / `identifier` / `version` / `out`。
- 全程无 cgo（仓库 `AGENTS.md` 硬规则 "No cgo"，原生调用走 purego），这是 §4 交叉编译成立的前提。

| 动作 | 命令 | 结果 |
|---|---|---|
| 单测 | `go test ./...` | `ok probe-native 0.550s` |
| 编译 | `go build -o probe-native.exe .` | 16,506,880 字节 |
| 启动 | `Start-Job { & .\probe-native.exe }` | 进程存活，窗口标题 `probe-native`，工作集 83.8 MB |

`mygo doctor`：go ✓、webview2 ✓（`157.0.4341.46`）、signtool ✓；bun ✗（前端模板需要，纯 Go 原生 UI 不需要）、makensis ✗（打 Windows 安装包时才下载 NSIS）。

## 3. moejs v0.1.1 (b7f49c1)：能嵌进 Go 进程并拿回返回值

探针 `tools/spike/moejs-eval`（`require github.com/MoeForge/moejs v0.1.1`）一次运行的全部输出：

```
compile OK
load OK
call summarize([1,2,3,4]) -> 10
call upper('abc') -> ABC @ host-time
boom -> error (expected): Error: kaboom
spin interrupt err: interrupted: context deadline exceeded
bench: 1000 calls in 998.7µs = 998ns/call
```

覆盖到的能力：`Compile` → `NewRuntime` → `SetGlobal`(NativeFunc) → `Load` → `Hook` → `Call` → `AppendJSON`；
JS 抛异常在 Go 侧是 `error` 而非 panic；死循环能被 `context` 超时 + `Interrupt` 打断并恢复；
调用开销约 1 µs/次（同机、同一 runtime 连调 1000 次）。

API 形态：ES module 插件模型（编译一次、可载入多个 runtime），面向"宿主调用插件导出函数"。它不是浏览器环境——没有 DOM、没有 `<video>`、没有 fetch（除非宿主显式给）。因此它可以做逻辑扩展点，不能承载播放器。

## 4. 交叉编译

```powershell
$env:CGO_ENABLED=0; $env:GOOS=darwin; $env:GOARCH=arm64; go build -o probe-native-darwin-arm64 .
```

| 目标 | 结果 | 产物 |
|---|---|---|
| windows/amd64（本机） | 成功 | 16,506,880 字节 |
| darwin/arm64（交叉） | 成功 | 15,472,194 字节 |

无 cgo，所以从 Windows 交叉编译 macOS arm64 的 Go 代码是通的。
但"编得过"不等于"能跑"：`mygo build` 产出 .app/.dmg、签名与公证仍需在 macOS 上完成。

## 5. 关键问题：目标 webview 里 MediaSource 在不在

CENC 在 Go 侧解密，webview 拿到的是代理吐出的明文分片，所以 webview 只需能喂 MSE（或原生 HLS）。这一条决定 B 路线成不成立。

### 5.1 Windows / WebView2（实跑）

探针 [tools/spike/README.md](../tools/spike/README.md) 里的 `webview-features`：Go 起本地 HTTP 服务 → mygo 开 webview → 页面做特性检测 → `fetch` 回报 → 落盘。

```json
{"ua":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/157.0.0.0 Safari/537.36 Edg/157.0.0.0",
 "hasMediaSource":true,"hasManagedMediaSource":false,"hasSourceBuffer":true,
 "mediaSourceType":"function","managedType":"undefined",
 "mp4Avc":true,"mp4Hevc":true,"nativeHls":"maybe",
 "newMediaSourceOK":true,"newMediaSourceErr":null}
```

- `MediaSource` 在，`isTypeSupported` 对 H.264 与 H.265 都为 true → hls.js 这条路可用。
- `ManagedMediaSource` 不存在（那是 iOS 的形态）。
- `canPlayType('application/vnd.apple.mpegurl')` 返回 `"maybe"`，不能当保证；播放实现仍应以 hls.js 为主。

### 5.2 macOS / WKWebView（源码依据，未实跑）

WebKit `main` 分支：

- `Source/WebCore/Modules/mediasource/MediaSource.idl`：`EnabledBySetting=MediaSourceEnabled`
- `Source/WebCore/Modules/mediasource/ManagedMediaSource.idl`：`EnabledBySetting=ManagedMediaSourceEnabled`
- `Source/WTF/Scripts/Preferences/UnifiedWebPreferences.yaml`：

```yaml
MediaSourceEnabled:
  status: embedder
  defaultValue:
    "ENABLE(MEDIA_SOURCE) && PLATFORM(IOS_FAMILY)": WebKit::defaultMediaSourceEnabled()
    "ENABLE(MEDIA_SOURCE) && !PLATFORM(IOS_FAMILY)": true
ManagedMediaSourceEnabled:
  status: mature
  condition: ENABLE(MEDIA_SOURCE)
  defaultValue:
    default: WebKit::defaultManagedMediaSourceEnabled()
    WebCore: true
```

读法：**非 iOS（含 macOS）的 MediaSource 默认开**，iOS 上才由 embedder 决定（所以 iOS 才需要 ManagedMediaSource 这个替代品）。

但这只是"默认值"，**必须真机实跑确认**：

1. `status: embedder` 意味着宿主（WKWebView 的 WKPreferences）可以覆盖它；
2. mygo 的 darwin 后端有没有动过相关 preference，我还没读；
3. macOS 自带原生 HLS，实现很可能绕开 MSE，走 MSE 的真实表现只能实测。

## 6. A 段：本地 HLS（fMP4 / TS）→ 代理 → hls.js → MSE 实测（Windows）

不碰外网、不碰真实源。做法：ffmpeg 造 30 秒 HLS（1280×720、H.264 + AAC、4 秒分片共 8 片），Go 进程用 `net/http` 提供页面、播放列表与分片，mygo 开 webview 加载页面，hls.js 1.6.15 播放并回报指标。第 4、5 组把播放列表和分片换成 `/proxy/` 路径：代理改写 m3u8 的每个 URI，分片按 8 KiB 块转发、每块节流 200 µs，模拟产品里"代理吐明文流"这一层。

ffmpeg 命令（**在分片目录内执行、必须用相对路径**，原因见 §7）：

```powershell
ffmpeg -y -f lavfi -i "testsrc2=size=1280x720:rate=30" -f lavfi -i "sine=frequency=440:sample_rate=48000" `
  -t 30 -c:v libx264 -preset veryfast -pix_fmt yuv420p -g 60 -keyint_min 60 -sc_threshold 0 `
  -c:a aac -b:a 128k -f hls -hls_time 4 -hls_playlist_type vod -hls_segment_type fmp4 `
  -hls_fmp4_init_filename "init.mp4" -hls_segment_filename "seg%d.m4s" "index.m3u8"
```

三次运行（WebView2 `157.0.4341.46`、hls.js `1.6.15`，页面与分片都走 127.0.0.1 回环）：

| 运行 | 首帧（MANIFEST_PARSED → playing） | seek 恢复（currentTime 赋值 → seeked） | 分片 |
|---|---|---|---|
| 1 | 371 ms | 34 ms | 8 片全部 200 |
| 2 | 152 ms | 28 ms | 8 片全部 200 |
| 3 | 183 ms | 32 ms | 8 片全部 200 |

再补两组对照（同一素材、同一页面）：

| 运行 | 首帧 | seek 恢复 | 说明 |
|---|---|---|---|
| 4：fMP4 经代理转发 | 610 ms | 35 ms | 播放列表被改写成代理自己的 URL，分片按 8 KiB 块转发、每块 200 µs 节流 |
| 5：**TS 分片**经代理转发 | 693 ms | 3 ms | 同一段视频改用 MPEG-TS 分片（`-hls_segment_type mpegts`），没有 `init.mp4` |

第 4 组说明：代理这一层只是"改写播放列表 + 分块转发"，就能把首帧从 152~371 ms 抬到 610 ms，**代理的实现质量直接决定首帧**，这条比 MSE 本身更值得盯。

第 5 组是容器格式对照组，结论重要：**hls.js 自己把 TS 转封装成了 fMP4 再喂 MSE**，所以 WebView2 的 MSE 全程没见过 TS，播放、seek 都正常（`directFetch` 404 是因为 TS 分片本来就没有 `init.mp4`，属预期）。
也就是说"TS 进不了 MSE"这个约束**不落在我们的代理层上，而落在 hls.js 上**：只要走 hls.js，代理吐 TS 还是 fMP4 都能播；只有在放弃 hls.js 改走原生 HLS（Safari 直接吃 m3u8）时，容器格式才重新变成问题。这一条对 macOS 的人天估算有直接影响，但 Safari 上 hls.js 的实际表现仍未验证（见 §9）。

内存（第 3 次运行，播放中 2.6 s 采样，进程树 7 个进程）：工作集合计 **315.7 MB**，私有内存合计 **177.9 MB**；
另一次 6 s 采样工作集合计 513.5 MB（8 个进程）。工作集跨进程重复计算共享页，**口径应以私有内存为准**。

**这些数字只能证明链路通，不能当达标证据**：回环没有真实网络抖动、没有 CENC、没有代理转封装、只有 30 秒素材、也没跑 30 分钟。首帧 ≤2 s 在这套环境里不构成任何结论，真实网络下它会由代理与源决定。

观察到的异常：fMP4 每次运行都有一条 `mediaError/bufferStalledError`（非 fatal），出现在 seek 之后到播放结束之间，不影响 `seeked` 与恢复；TS 那次另有一条 `bufferSeekOverHole`。初判为短素材 + VOD 末尾缓冲边界，需在长播里复核。

## 7. 这次踩到并已定性的坑

1. **ffmpeg 把绝对路径写进播放列表**：`-hls_fmp4_init_filename` 给绝对路径时，`#EXT-X-MAP:URI="C:\...\init.mp4"` 会原样写进 m3u8；hls.js 把它当成 Windows 文件路径去请求，浏览器里表现为 6 条 `networkError/fragLoadError`（`code=0`、`url=''`），而 HTTP 服务器**一条分片请求都收不到**。
   排查方式：给 hls.js 的 `Events.FRAG_LOADING` 打点，直接打 `frag.url`。生成脚本里必须用相对路径。
2. **WebView2 控制器 `0x8000FFFF`**：受限会话里创建 WebView2 连续失败三次；脱离该会话启动一次成功。与 mygo 无关，但 CI、服务化、受限终端都会遇到。
3. **`enableWorker` 不是原因**：关掉 worker 后行为完全一致，问题不在 worker。
4. **`mygo init` 模板不锁版本**：模板用 `cmd/mygo@latest`，要钉死必须显式 `go get -tool …@v0.9.2`。

## 8. 结论

1. **工具链能落地**：Go 1.27.4 + mygo v0.9.2 (25ce62f) 能起窗口、能跑单测，能从这台 Windows 机器交叉编译出 darwin/arm64 二进制。
2. **moejs 能用，但只适合当逻辑扩展点**：嵌入、调用、异常隔离、超时中断都成立，约 1 µs/次；但它没有 DOM/video，承载不了播放；v0.1.1 (b7f49c1) 只有两个正式版、公开生产案例为 0。
3. **B 路线在 Windows 上验到"能播"**：WebView2 有 MediaSource，本地 fMP4 HLS 经 hls.js 播放、seek 均成功；再套一层改写播放列表、分块转发的代理，首帧 610 ms、seek 35 ms，链路仍然通。
4. **容器格式的压力不在代理层**：TS 分片经同一代理、同一 hls.js 也能播（首帧 693 ms、seek 3 ms），因为 hls.js 在 JS 里把 TS 转封装成 fMP4 才交给 MSE。只要走 hls.js，"代理必须转封装成 fMP4"这一笔工作量可以不计；改走原生 HLS 时才需要。
5. **B 路线在 macOS 上只验到"源码默认开"**：MediaSource 默认开，但 WKWebView 实跑未做。
   这是当前唯一能决定 B 路线生死、而我在这台机器上无法闭合的验证。

## 9. 未覆盖

- **macOS 真机**：特性检测、首帧、seek、内存全未测。`tools/spike/` 的三个探针在 Mac 上重跑即可（`go build` + 直接运行，无需 Bun）。
- **真实源与对照组**：按约定不在公司设备执行（B 段，用户个人设备，30 分钟 ×3 源 ×2 平台）。
- **CENC 解密 → 代理 → MSE** 全链：本地素材没有加密，Go 侧解密与转封装未参与。
- **代理本身的吞吐与延迟**：A 段分片直接从本地文件服务发出，不含代理实现。
- **30 分钟长播的内存曲线**：只有 30 秒素材的 2 次采样。
- **打包链路**：`mygo build` 的 .app/.dmg、签名、公证、自动更新均未验证（需 Mac 与证书）。
- **hls.js 版本选型**：只测了 1.6.15，未对比其他版本。
