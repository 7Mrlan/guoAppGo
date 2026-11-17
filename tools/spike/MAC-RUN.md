# 在 Mac 上跑一次（10 分钟）

目的只有一个：把 macOS 的 `MediaSource`、hls.js 播放、moejs 嵌入这三件事从"源码推断"变成"实测"，销掉 `docs/spike-stack.md` §5.2、§9 里的挂账。**不要在这页之外另搭播放器。**

## 0. 前提

- macOS 14 或更高、Apple silicon。
- Go 1.27.4（`go version` 能看到即可）。
- `ffmpeg`（`brew install ffmpeg`；只在造测试素材时用一次）。
- 全程只需本机：不访问任何视频源、不碰解密、不装 Flutter。

## 1. 拿代码

```sh
git clone --branch develop https://github.com/7Mrlan/guoAppGo.git
cd guoAppGo/tools/spike
```

## 2. 编出三个探针

在这台 Mac 上直接编（最稳）：

```sh
(cd moejs-eval       && go build -o /tmp/moejs-eval .)
(cd webview-features && go build -o /tmp/webview-features .)
(cd mse-chain        && go build -o /tmp/mse-chain .)
```

也可以从 Windows 交叉编译后拷过来（`CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o mac-<名字> .`），但**优先在本机编**：本机编能排掉"交叉编译产物不可运行"这一类问题。

## 3. 特性检测（1 分钟，最要紧的一条）

```sh
PROBE_OUT=/tmp/features.json /tmp/webview-features
cat /tmp/features.json
```

期望 JSON 里这三项：`hasMediaSource`、`mp4Avc`、`nativeHls`。判读：

| 字段 | 通过 | 不通过的含义 |
|---|---|---|
| `hasMediaSource` | `true` | `false` → hls.js 在主平台不成立，B 路线要改架构 |
| `mp4Avc` / `mp4Hevc` | `true` | `false` → 解码器不支持该编码，源要转码 |
| `nativeHls` | `maybe`/`probably` | 空串 → 原生 HLS 兜底没有，只能走 MSE |
| `hasManagedMediaSource` | 不判 | macOS 没有是正常的（它是 iOS 的形态） |

## 4. moejs 嵌入（1 分钟）

```sh
/tmp/moejs-eval
```

期望 6 行：`compile OK`、`load OK`、`call summarize([1,2,3,4]) -> 10`、`call ABC @ host-time`、`boom -> error`、`spin interrupt err:`，最后一行 `bench: … = 约 1µs/call`。

## 5. 播放链路（3 分钟）

```sh
export PROBE_DIR=/tmp/probe-mac
mkdir -p $PROBE_DIR/hls $PROBE_DIR/web
cd $PROBE_DIR/hls
ffmpeg -y -f lavfi -i "testsrc2=size=1280x720:rate=30" -f lavfi -i "sine=frequency=440" \
  -t 30 -c:v libx264 -preset veryfast -pix_fmt yuv420p -g 60 -keyint_min 60 -sc_threshold 0 \
  -c:a aac -b:a 128k -f hls -hls_time 4 -hls_playlist_type vod -hls_segment_type fmp4 \
  -hls_fmp4_init_filename "init.mp4" -hls_segment_filename "seg%d.m4s" "index.m3u8"

# hls.js：版本必须与文档一致（1.6.15）
curl -L -o $PROBE_DIR/web/hls.min.js https://cdn.jsdelivr.net/npm/hls.js@1.6.15/dist/hls.min.js

PROBE_DIR=$PROBE_DIR /tmp/mse-chain
cat $PROBE_DIR/mse-chain.json
```

**分片必须在 `hls/` 目录内用相对路径生成**，否则 ffmpeg 会把绝对路径写进 m3u8 的 `EXT-X-MAP`，hls.js 直接报 `fragLoadError`（`docs/spike-stack.md` §7 第 1 条）。

## 6. 如果裸二进制起不来

WKWebView 的宿主偏好按应用标识生效，裸二进制可能拿不到真正 .app 的行为。改走这条路：

```sh
mkdir -p /tmp/mse-mac && cd /tmp/mse-mac
go mod init probe/macrun
go get github.com/egoist/mygo@v0.9.2
# REPO 换成第 1 步 clone 下来的仓库路径
cp "$REPO/tools/spike/mse-chain/main.go" .
go build -o /tmp/mse-chain .
go tool mygo version
```

也可以把 `main.go` 放进一个 `mygo init -template native` 生成的工程里，用 `go tool mygo dev` 跑——它会产出带应用标识的 Dev 应用包，最接近真实运行形态。两条都试不通时，把报错原文贴回来，这本身就是结论。

## 7. 结果怎么回传

贴两样东西即可：

1. `/tmp/features.json`（完整原文）；
2. `$PROBE_DIR/mse-chain.json`（完整原文），以及 `sw_vers && sysctl -n machdep.cpu.brand_string`。

内存按**私有内存**记（与 R0 口径一致）：Activity Monitor → Memory 列，或 `footprint -p <pid>`；工作集只作参考。采样点在"播放中"和"seek 后 10 秒"各一次。

另外记一行**分片容器类型**：有 `init.mp4`、分片是 `.m4s`/`.mp4` 即 fMP4，分片是 `.ts` 即 MPEG-TS。原因见 `docs/spike-stack.md` §6 第 5 组：hls.js 会把 TS 在 JS 里转封装成 fMP4，所以在 hls.js 路线上容器不影响播放；但若改走原生 `<video>` HLS（Safari 直接吃 m3u8），TS 的处理就重新变成工作量。真实源是哪种，只有 B 段能回答。

## 8. 判读标准

| 项 | 通过线 | 说明 |
|---|---|---|
| `hasMediaSource` | `true` | 这条不过，B 在主平台不成立，别再测播放 |
| 首帧 `firstFrameMs` | ≤2000 | 本机回环，实际会远小于它；若判不过，先查是否走了代理节流 |
| seek `seekRecoverMs` | ≤1500 | 同上 |
| `errors` | 只允许非 fatal 的 `bufferStalledError` / `bufferSeekOverHole` | 出现 `fragParsingError`、`fatal` 先看容器与编码 |

**对照组不放这页**：原 Flutter 应用必须用真实源跑 30 分钟，按 `docs/testing/acceptance.md` 的表格人工记，且只在个人设备上做。
