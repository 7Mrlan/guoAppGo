# guoAppGo 技术评估（一）：版本基准与现状

评估：探路 · 数据采集时间 2026-11-17 00:05–00:15（America/Los_Angeles）· 所有结论以第 0 节版本基准为准

> 本文只覆盖"基线、原应用规模、MyGo、moejs"；**路线取舍、工作量估算、风险清单在 [evaluation-plan.md](evaluation-plan.md)**（两文按章节边界拆分，各自独立可读，章节编号连续不重起）。

---

## 0. 版本基准

| 项目 | 仓库 | tag | commit sha | commit 日期 (UTC) |
| --- | --- | --- | --- | --- |
| 原应用 小果影视 | Xanerix/guoapp | `app-v2.0.1-113` | `4c7aab6a87a12e1aaf6b1d08fa31d3844148650c` | 2026-11-16T16:21:19Z |
| 桌面框架 | egoist/mygo | `v0.9.2` | `25ce62f4d357cd6d29da04300e3fbc352d79daa4` | 2026-11-16T12:54:29Z |
| JS 运行时 | MoeForge/moejs | `v0.1.1` | `b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070` | 2026-11-15T17:05:00Z |

与 [versions.md](versions.md) 一致（该文件为唯一事实来源，本表仅复述）。本文所有行数、文件数、issue 数、release 数均在这三个 tag 的代码上实测，命令见附录 A。

**tag 与 main 的差异**（`git diff --shortstat app-v2.0.1-113..5a7a116`）：

- 全仓 83 个文件、+2,058 / −1,032；其中 Dart 20 个文件、+1,291 / −430。main 版本号为 `2.0.2+114`。
- 改动集中在弹幕表情（新增 `lib/danmaku_emoji.dart`）、`lib/playback_session.dart`、竖屏播放器（`portrait_player_page.dart` +214）与详情页季条（`season_strip.dart`，新增）。
- `lib/ui/detail/detail_page.dart` 是 **+97/−315，净删 218 行**——main 把详情页拆了一块出去（内容挪进 `season_strip.dart`）。所以按 main 估会把规模**算少**，而不是算多；这也是本文坚持按 tag 估的原因。
- `native/` 完全一致（diff 0 文件），所以**以 tag 估 Go 侧工作量不会漏**；但 2.0.2 的弹幕表情（约 210 行）不在基线内，重写范围按 tag 算，别按 main 算。
- `packages/` 10 文件 +493/−598：`media_kit_video/windows` 把渲染后端从 ANGLE 换成 D3D11（删 `angle_surface_manager.*`，新增 `d3d11_surface_manager.*`、`render_d3d11.h`）。我们不用 media_kit，实现不受影响；但这说明上游在 Windows 渲染上踩过坑，Windows 附加端别自己再走 ANGLE。

---

## 1. 结论（先看这段）

1. **建议做。路线倾向 B（Go 核心 + webview 前端，播放器页也放 webview），但最终 A/B 等 R0 出结果再锁**：B 的优势全押在"播放能塞进 webview"上，这条未验证前不应当成既定事实。
2. **前提纠正：这次不是"把 54k 行 Dart 翻译成 Go"。** `native/` 已有 **13,116 行 Go（非测试 11,784 行、77 个文件）、47 个 JSON RPC op**，站源、取流、下载抓取、CENC 解密、投屏、缓存、设置、更新都在里面；Dart 侧只通过一个 C ABI 入口（`DuanjuRequest`/`DuanjuFree`）调用。要做的是：**给现成核心配 UI 壳 + 平台层 + 进程模型**。
3. **真正的技术缺口是两处，且都不在 Go 里**：① 播放器——`media_kit`（libmpv）1.2.6 + `lib/playback_session.dart` 1,397 行解码器/轨道/硬解编排；② 转码——`ffmpeg_kit` 在 Dart 侧做转封装、m3u8 合并、probe、封面解码，**`native/` 里 `ffmpeg|exec.Command` 命中 0**。这两块决定工期量级，UI 只决定日历长度。
4. **现实工作量（v1 单源主口径）：B 106 人天、A 133 人天、C 115~125 人天**（含 10% 缓冲，口径见 [evaluation-plan.md](evaluation-plan.md) 6.1）。2 人并行约 **12.5 周**（B）。多源口径为 B 117 / A 142；**B 是条件值：R1 不通过则单源 126~136，与 A 持平**（见同文 R0）。开工前先做 4 个 spike，共 **5 人天**（同文第 7 节）。
5. **moejs 已出局（v1 单源已定）**：Chief 2026-11-17 拍板 v1 只做红果——模块系统整层不做，goja 那约 880 行 JS 桥不替换，moejs 不进 `go.mod`，ADR-0009 不写。多源放 v2 再评估，理由与代价见 4.5。
6. 两个要拍板的事：**Windows 上 HEVC/AV1 是否必须全量可播**（决定要不要嵌 mpv，+20~30 人天）、**是否接受 GPLv3 源码义务**（原仓库就是 GPLv3，见 2.5）。

---

## 2. 原始规模（`app-v2.0.1-113` 实测）

### 2.1 代码量

| 范围 | 文件数 | 行数 | 备注 |
| --- | ---: | ---: | --- |
| Dart 合计（含 `packages/`、`test/`） | 192 | 47,460 | 全仓 `*.dart`，排除 `build/`、`.dart_tool/` |
| `lib/` | 132 | 37,219 | 应用本体 |
| `lib/ui/` | 89 | 27,822 | 占 `lib/` 的 74.7% |
| `lib/ui/player/` | 19 | 8,891 | 播放器、弹幕层、选集、迷你播放器、锁 |
| `lib/ui/{settings,downloads,modules}/` | 18 | 6,408 | 设置/下载/站源模块页 |
| 弹幕相关（4 个文件） | 4 | 1,326 | controller/models/style/overlay |
| 非 UI Dart（`lib/` 根下） | 43 | 9,397 | 存储、编排、平台适配 |
| Go `native/` 合计 | 87 | 13,116 | 含测试 |
| Go `native/` 非测试 | 77 | 11,784 | 可直接复用的业务核心 |
| Dart 测试 | 28 | — | 仓库自带 `test/` |

**页面数：19 个** `lib/**/*_page.dart`（详情、探索、搜索、追剧、历史、我的、下载、下载选择、已下载剧、设置、站源模块、投屏、竖屏/横屏/电视播放、播放、解锁、条款）。
页面之外还有大量非页面组件：`lib/ui/widgets/` 19 个文件（卡片、网格、搜索框、进度条、字母键盘等）。

### 2.2 状态管理

- `pubspec.yaml` 里**没有任何状态管理库**：无 provider、riverpod、bloc、getx。
- 代码文本匹配计数：`setState` 275 处、`ChangeNotifier` 15 处、`ValueNotifier` 13 处、`InheritedWidget` 11 处。
- 结论：状态层是 Flutter 原生的 `setState` + 手写 `AppScope`（`lib/ui/app_scope.dart`）。迁移时**不需要翻译状态管理框架**，需要做的是把 9,397 行非 UI Dart 的职责并回 Go core。

### 2.3 依赖清单（21 个直接依赖 + 3 个本地覆盖）

| 类别 | 依赖 | 版本 |
| --- | --- | --- |
| 播放 | media_kit / media_kit_video / media_kit_libs_video | 1.2.6 / 2.0.1 / ^1.0.7 |
| 播放 | video_player | ^2.9.2（`lib/` 内无引用，Android 侧残留） |
| 转码 | ffmpeg_kit_flutter_new_min_gpl | 2.6.7（**GPL 构建**） |
| 存储 | shared_preferences / shared_preferences_windows（本地覆盖） | ^2.5.5 / path |
| 存储 | path_provider / path_provider_windows / path_provider_platform_interface | ^2.1.5 / ^2.3.0 / ^2.1.3 |
| 系统 | window_manager / package_info_plus / url_launcher / file_picker | ^0.5.1 / ^10.2.1 / ^6.3.3 / 13.1.0 |
| 后台 | flutter_foreground_task | 11.0.3（Android 后台下载/听书） |
| 主题 | dynamic_color | ^2.1.0 |
| 其他 | ffi / crypto / lpinyin | ^2.2.0 / ^3.0.7 / ^2.0.3 |

本地覆盖：`media_kit_video`、`media_kit_libs_windows_video`、`shared_preferences_windows`（`packages/` 下自带补丁）。
**无国际化**：仓库内 `.arb` 文件 0 个，文案硬编码中文；重写不背 i18n 债。

### 2.4 架构事实：Go 核心已经存在

- `native/bridge/main.go` 只导出两个 C 函数：`DuanjuRequest(char*) -> char*`（JSON 进、JSON 出）与 `DuanjuFree`。Dart 侧 `lib/core_bridge.dart` 1,139 行做封装。
- `native/core/` 77 个非测试文件，覆盖：`app_sources`（站源）、`app_playback`/`app_stream`（取流、CENC 解密、本地流代理）、`app_downloads`（**只做 HLS 分片抓取、命名、合集与迁移**）、`app_cast`（投屏）、`app_catalog_cache`、`app_settings`、`app_update`、`module_*`（goja 站源模块宿主）。
- **合并与转码不在 Go 里**：`native/` 全仓 grep `ffmpeg`、`exec.Command` 命中 0；mp4 转封装、m3u8 合并、时长 probe、封面解码都在 Dart（`lib/media_pipeline.dart`、`lib/download_converter.dart`、`lib/cover_decoder.dart`）。重写必须自带 ffmpeg（随包分发 CLI 或纯 Go remux），这笔费在 M5，风险在 R5。
- **Go 现在是"库"不是"主程序"**：`app_runtime.go` 的 dispatch 有 47 个 op，其中 `release`、`cancelPlayback`、`workLease` 是给 Flutter 做优雅关停用的。重写后 Go 成为主进程，启动/退出、任务租约、播放取消的生命周期要重做（M13），不能只做接口搬家。
- 站源 JS 桥的替换面：`module_script.go`（475 行，建 VM、CommonJS 包装、JSON 编解码）+ `module_host.go`（406 行，`host` 对象：`fetch`、`exclusive` 等）≈ 880 行；其下 `provider_*.go`（1,545 行）与 JS 引擎无关。**换引擎不动业务层**，风险集中在 `fetch` 的超时、限流、重定向、状态码白名单语义是否等价。
- 含义：重写后可以 `mygo.NewWindow(...)` 直接调 `core.NativeRequest(...)`，**不再有 FFI + JSON 边界**；"站点逻辑放 JS 模块"的设计可以原样保留（若保留多站源）。

### 2.5 许可证（影响发布方式）

原仓库 `LICENSE` 为 **GNU GPL v3**，且依赖 `ffmpeg_kit_flutter_new_min_gpl`（GPL 构建）。若 guoAppGo 复用原 Go 代码或继续捆绑 ffmpeg，发布物同样受 GPLv3 约束（分发时须提供源码）。若要做闭源分发，必须重写被复用的部分并移除 GPL 组件——这是**产品决策**，不是技术细节，已列入风险 R5。

---

## 3. MyGo 现状（`v0.9.2`，`25ce62f`）

### 3.1 项目活跃度（GitHub API，2026-11-17 00:05）

| 指标 | 数值 |
| --- | --- |
| star / fork | 6,448 / 406 |
| 创建时间 | 2026-09-26 |
| release 总数 | **83 个，7 周**（最近 `v0.9.2`，2026-11-16） |
| 提交密度 | `v0.9.0`→`v0.9.2`：5 天 146 commits、300 个文件 |
| 测试文件 | 854 个 `*_test.go`，其中 `ui/` 309 个 |
| 贡献者 | 100（API 单页上限，实际可能更多） |
| 未关闭 issue | 约 139（含 PR） |

判断：**活跃度极高，但 API 仍在快速演化**。83 个 release 里有 12 个在 release notes 中提及 breaking（`0.2.13`、`0.3.0`、`0.4.0`、`0.5.0`、`0.7.0`、`0.7.2`、`0.8.0`、`0.9.0` 等）；`0.9.1`/`0.9.2` 明确写 "No breaking changes to the Go API"，`0.9.0` 的 breaking 仅限实验性 native modules。

### 3.2 两条 UI 路线成熟度

| 维度 | 原生 UI（`ui` 包） | webview |
| --- | --- | --- |
| 技术 | GPU 直绘，Go 写视图函数 | WKWebView（macOS）/ WebView2（Windows）/ WebKitGTK |
| 体量 | **272 个以 `*ui.Context` 为首参的构造函数**（含 `coreXxx` 内部变体） | 41 个相关 issue |
| 布局 | flexbox + grid、虚拟列表 `ui.List(c, state, n, row)` | 浏览器布局 |
| 文本 | 编辑、IME、选区、无障碍 | 浏览器原生 |
| 媒体 | `ui/video`：系统解码、HLS、GPU 帧 | `<video>`：HLS（Safari 原生）、MSE |
| 画中画 | **仅页面内 `<video>`**（`WindowOptions.PictureInPicture`，`page.go TogglePictureInPicture`） | 原生支持 |
| 学习成本 | 高（0.9.x、文档 668 行 `docs/native.md`、每周在变） | 低（前端栈） |
| 测试 | 309 个 ui 测试，支持无窗口视图测试 | TypeScript client 自动生成 |

结论：**webview 路线成熟；原生 UI 路线"能力齐、资历浅"**。原生 UI 的 272 个控件覆盖了做一个桌面影视应用所需的大部分（列表、视频、滑块、弹层、菜单、上下文菜单、输入框、颜色选择器、日历），但播放器周边（画中画、滚动条、触摸拖动）仍在补齐。

### 3.3 平台能力矩阵（对 guoapp 的可用性）

| 能力 | v0.9.2 支持 | 证据 | 对 guoapp |
| --- | --- | --- | --- |
| 菜单 / 快捷键 | 是 | `menu.go`、`GlobalShortcut` 57 处 | 可用 |
| 托盘 / 菜单栏 | 是 | `Tray` 613 处、`examples/menubar-*` | 可用（迷你播放器入口） |
| 通知 | 是 | `notifications_test.go`、`docs/native.md` | 可用 |
| 文件关联 / 深链 / 最近文档 | 是 | `appfile.go`、`deeplink.go`、`FileAssociation` 38 处 | 可用 |
| Now Playing / 媒体键 | 是 | `nowplaying.go` 164 处 | 可用（播放页必需） |
| 自动更新 | 是，签名 + 增量（Sparkle 式） | `updater.go`、`docs/updates.md` | 可用，替换 `app_update.go` 的下载逻辑 |
| 签名 / 公证 / 安装包 | 是 | `cmd/mygo/codesign.go`、`Notarize` 27 处、`docs/distribution.md` | 可用（需开发者账号） |
| 窗口 | 无边框、vibrancy、圆角、置顶、Panel、全屏 | `window.go` 30+ 选项 | 比 window_manager 更强 |
| 视频解码 | AVFoundation / Media Foundation；H.264+AAC 必过，HEVC/VP9/AV1 看系统；HLS 在 mac/win 支持 | `ui/video/video.go` 包注释 | 够用，但**无 libmpv 兜底** |
| 自定义请求头（Referer/UA） | **无** | `video.Open/OpenURL/OpenData` 无 header 参数 | 用现有 `app_stream.go` 本地代理补 |
| 画中画（原生视频） | **无** | PiP 仅页面视频 | 原生 UI 路线缺少此能力 |
| 投屏 / DLNA / Chromecast | **无** | 全仓 grep 0 命中 | 复用原 Go `app_cast.go` 自研 |
| 多用户 | 无 | — | 自研（原实现也简单） |

### 3.4 已知坑（真实 issue，均已核对）

| issue | 内容 | 状态 |
| --- | --- | --- |
| #1743 | macOS 上 `video.OpenURL` 时 AVFoundation 加载导致应用冻结 | 2026-11-12 已修 |
| #1830 | macOS 上 `ui.Video` 首帧不出 | 2026-11-14 已修 |
| #1752 | Windows 上含非 ASCII 字符路径 `video.Open` 失败 | 2026-11-12 已修 |
| #973 | 零拷贝 GPU 纹理互操作（FFmpeg VideoToolbox / D3D11） | open |
| #1712 | 触摸屏单指拖动应滚动（Windows/Linux） | open |
| #1776 | 滚动时显示 overlay 滚动条 | open |
| #1796 | 在 Linux 上构建并签名的 macOS 应用自动更新架构问题 | open |

判断：**修得快（多为当天到 3 天内）**，但恰好集中在视频、滚动、更新这三块我们要重度使用的地方，说明这些能力刚在打磨期。

### 3.5 与 Flutter 的能力差距

| 差距 | 影响 | 缓解 |
| --- | --- | --- |
| 无内置 libmpv | HEVC/AV1 在 Windows 依赖系统扩展；异常封装格式无软件兜底 | 本地代理 + 明确降级策略；必要时 native module 嵌 mpv（+20~30 人天） |
| 原生 UI 无画中画 | 迷你播放器/画中画要手写 | 选 webview 路线直接消掉 |
| 无投屏 | 需自研设备发现/遥控 | 原 `app_cast.go` 是 Go，可直接复用 |
| 控件库无"设计系统" | 圆角、动效、毛玻璃要自己搭 | `glass` 插件 + 统一主题层 |
| 要求 Go ≥ 1.27.3 | 构建机必须升级 | 本机当前未装 Go（见戈的 spike） |

---

## 4. moejs 现状（`v0.1.1`，`b7f49c1`）

### 4.1 它是什么（先说清定位）

- **纯 Go 实现的 JavaScript 运行时**，代码 **251,000 行 Go**，不是一个 UI 框架，也不是插件框架；官方定位是"在 Go 程序里跑 JS 插件"。
- README：通过 **79,385 条 test262 用例**（ES modules、classes、async/await、Proxy、BigInt）。
- 自带 `eventloop`（setTimeout/setInterval）、`webapi`（fetch/Headers/Streams/AbortController/URL）、`node`（`node:fs`、`node:buffer`，用 `os.Root` 限定目录）。

### 4.2 API 形态（从 v0.1.1 源码读出）

| 层 | API |
| --- | --- |
| 运行时 | `NewRuntime(Options)`、`Options.MemoryLimit`、`Stats` |
| 编译 | `Compile(name, src)`/`CompileTS` → `*Module`；`Link(entry, Resolver)`；`Module.Exports()`/`Requests()`/`Hook(export, members...)` |
| 脚本 | `CompileScript`、`CompileFunction`、`Script.Run` 类接口 |
| 值 | `Value`、`Promise`、`Unmarshal`/`DecodeOptions`（结果可直接进 struct，等价 `json.Unmarshal`） |
| 宿主函数 | 普通 Go 函数；可返回 promise，由 Go 侧异步 settle |
| 错误 | `*SyntaxError`、`*InternalError`，可中断死循环 |

**限制（会影响 guoapp 现有模块）**：

1. **只支持 ES Module，不支持 CommonJS**（全仓 grep `module.exports`/`require(` 命中 0）。而 guoapp 站源模块是 `module.exports = {...}` 形态——需要一层 wrapper（把 `module`/`exports` 注入脚本作用域），约 1~2 人天。
2. 不支持 JSON modules / import attributes、decorators、`using`、`Temporal`、Iterator helpers（`TODO.md` 明列）。
3. 依赖 `github.com/MoeForge/moego v0.1.0-alpha.3`（**alpha**），Go ≥ 1.25。

### 4.3 性能口径

README 实测（2026-10-01 采集，new-api 插件负载）：

| 指标 | moejs | Sobek | QuickJS(modernc) | quickjs-go |
| --- | ---: | ---: | ---: | ---: |
| 单次插件调用 | **6.9 µs** | 14.3 µs | 35.2 µs | 104.5 µs |
| 新建 runtime | **1.4 µs** | 2.2 µs | 180 µs | 382 µs |
| 每 runtime 内存 | **81 KiB** | 264 KiB | 227 KiB | 348 KiB |

注意：这是"插件调用"负载，**不是**影视站源的 HTML 解析 + 正则抓取；且 README **没有与 goja 的对比**（goja 正是 guoapp 当前使用的引擎），所以无法直接量化"换 moejs 能快多少"。

### 4.4 成熟度（回答 Chief 的三问）

| 问题 | 证据 | 判定 |
| --- | --- | --- |
| API 稳定性 | 正式版只有 `v0.1.0`（11-09）与 `v0.1.1`（11-16）；两版之间 7 commits / 85 文件，release notes 无 breaking 声明 | 0.x 语义上**不承诺稳定**，但目前未见破坏性变更 |
| 生产案例 / 活跃度 | 712 star、6 位贡献者、15 个 issue 全部 closed（中位 20 小时）、7 周 4 个 release | 活跃，但**查不到公开生产案例**（README 未列） |
| 明天改 API / 停更的代价 | 站源模块层被 `module_host.go`（406 行）+ `module_script.go`（17 KB）包住；全仓 `goja.` 引用仅 51 处 | 可隔离，见下 |

**隔离方案**：定义 `script.Engine` 接口（`Execute(script) (Value, error)`、`RegisterHost(fn)`、`Call(export, args)`、`Close()`），goja 与 moejs 各一个实现；业务与站源模块不 import 任何引擎包。切换/回退成本约 **5~8 人天**。停更则回退 goja，业务零改动。

### 4.5 结论

- **v1 已定单源，moejs 出局**（Chief 2026-11-17 拍板，出处：本群当轮决定）：模块系统整层不做，goja 的约 880 行 JS 桥不替换，moejs 不进 `go.mod`、ADR-0009 不写。
- 若 v2 恢复多站源：默认实现仍是 goja（现有生产验证过的），moejs 作为第二实现，用"同一批站源模块在两种引擎下输出 diff 为 0"作为切换门槛（见 R4）。
- 若要走第二步，先把它放进 `script.Engine` 接口后，不让 moejs 的 `Value`/`Runtime` 类型泄漏到业务代码，避免第三次迁移。

---

## 附录 A：复现命令

```powershell
# 工作区（本机 D 盘不在沙箱内时的临时目录）
$ws = "C:\Users\1009638\DeepSeek Bot\_eval"; cd $ws

# 1) 按 tag 拉取三个仓库
git clone --quiet https://github.com/Xanerix/guoapp.git
cd guoapp
git fetch --quiet --depth 1 origin refs/tags/app-v2.0.1-113:refs/tags/app-v2.0.1-113
git rev-parse app-v2.0.1-113          # 4c7aab6a87a12e1aaf6b1d08fa31d3844148650c
git rev-parse HEAD                    # 5a7a116...（main，仅用于对比）
git worktree add ..\guoapp-201 app-v2.0.1-113

git clone --quiet --branch v0.9.2  --depth 1 https://github.com/egoist/mygo.git  mygo
git clone --quiet --branch v0.1.1  --depth 1 https://github.com/MoeForge/moejs.git moejs
git -C mygo   rev-parse HEAD          # 25ce62f4d357cd6d29da04300e3fbc352d79daa4
git -C moejs  rev-parse HEAD          # b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070

# 2) 规模统计（在 guoapp-201 下）
$all = Get-ChildItem -Recurse -Filter *.dart -File | Where-Object { $_.FullName -notmatch '\\build\\|\\\.dart_tool\\' }
$all.Count; ($all | Get-Content | Measure-Object -Line).Lines      # 192 / 47460
$lib = Get-ChildItem lib -Recurse -Filter *.dart -File
$lib.Count; ($lib | Get-Content | Measure-Object -Line).Lines      # 132 / 37219
(Get-ChildItem lib -Recurse -Filter *_page.dart).Count             # 19
$g = Get-ChildItem native -Recurse -Filter *.go -File
($g | Where-Object { $_.Name -notmatch '_test\.go$' }).Count        # 77
($g | Where-Object { $_.Name -notmatch '_test\.go$' } | Get-Content | Measure-Object -Line).Lines  # 11784

# 3) 状态管理 / 依赖（文本匹配计数，非 AST）
foreach ($p in 'ChangeNotifier','ValueNotifier','setState','InheritedWidget') {
  "$p : " + (Get-ChildItem lib -Recurse -Filter *.dart | Select-String -Pattern $p -SimpleMatch | Measure-Object).Count
}

# 4) MyGo / moejs 项目指标（GitHub REST API，经 node 的 TLS 通道；本机 PowerShell 的 HTTPS 不通）
node gh.js      # repo stars/forks/releases/issues
node gh2.js     # 破坏性变更提及、版本间 diff、issue 关闭中位时长
```
