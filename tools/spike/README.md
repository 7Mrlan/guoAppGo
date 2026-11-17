# spike：栈可行性验证的最小程序

三个探针，对应 `docs/spike-stack.md` 的 §3、§5、§6。都不属于产品代码，只为让结论可复现。

| 目录 | 用途 | 产出 |
|---|---|---|
| `moejs-eval` | moejs v0.1.1 嵌入 Go、调用 JS、异常与超时 | stdout |
| `webview-features` | 系统 webview 里的 MediaSource / ManagedMediaSource / 编解码支持 | `PROBE_OUT` 指定的 JSON |
| `mse-chain` | 本地 HLS → 代理（改写播放列表、分块转发）→ hls.js → MSE，量首帧与 seek；fMP4 与 TS 两种分片都能跑 | `<PROBE_DIR>/mse-chain.json` |

版本钉死：`mygo v0.9.2`（`25ce62f`）、`moejs v0.1.1`（`b7f49c1`）、Go 1.27.4。

## 跑法

```powershell
go build -o probe.exe .
```

`moejs-eval` 直接运行即可。

`webview-features`：

```powershell
$env:PROBE_OUT = "C:\tmp\features.json"
.\probe.exe
Get-Content $env:PROBE_OUT
```

`mse-chain` 需要先准备素材与 hls.js：

```powershell
# 1) 分片：必须在目标目录内用相对路径执行，否则 ffmpeg 会把绝对路径写进 m3u8 的 EXT-X-MAP
cd <PROBE_DIR>\hls
ffmpeg -y -f lavfi -i "testsrc2=size=1280x720:rate=30" -f lavfi -i "sine=frequency=440:sample_rate=48000" `
  -t 30 -c:v libx264 -preset veryfast -pix_fmt yuv420p -g 60 -keyint_min 60 -sc_threshold 0 `
  -c:a aac -b:a 128k -f hls -hls_time 4 -hls_playlist_type vod -hls_segment_type fmp4 `
  -hls_fmp4_init_filename "init.mp4" -hls_segment_filename "seg%d.m4s" "index.m3u8"

# 2) hls.js（版本要与结论一致）
#    https://cdn.jsdelivr.net/npm/hls.js@1.6.15/dist/hls.min.js -> <PROBE_DIR>\web\hls.min.js

# 3) 运行
$env:PROBE_DIR = "<PROBE_DIR>"
.\probe.exe
Get-Content <PROBE_DIR>\mse-chain.json
```

`HLS_JS` 环境变量可以换用别的 hls.js 构建做对比。

## 机器相关的注意点

- 在受限会话（非用户桌面会话）里启动时，WebView2 会报 `0x8000FFFF`。换到桌面会话，或用 `Invoke-CimMethod Win32_Process Create` 拉起。
- go 的 `%TEMP%` / `%USERPROFILE%\go` 不可写的机器上，需要设置 `GOTMPDIR`、`GOCACHE`、`GOPATH`。
