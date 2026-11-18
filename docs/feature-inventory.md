# 原应用功能清单与优先级

本文回答三个问题：原应用有哪些功能块、每块该给什么优先级、重写时哪些可以砍或简化。
所有结论标注证据文件与行数，统计命令见 §1，任何人都能复现。

---

## 0. 版本基准

| 项目 | 版本 | commit sha | 说明 |
| --- | --- | --- | --- |
| guoapp（原应用） | `app-v2.0.1-113`（`pubspec.yaml` 为 `2.0.1+113`） | `4c7aab6a87a12e1aaf6b1d08fa31d3844148650c` | 2026-11-17 00:21:19 +0800；本文全部功能与规模结论的唯一依据 |
| egoist/mygo | `v0.9.2` | `25ce62f4d357cd6d29da04300e3fbc352d79daa4` | 2026-11-16；本文不评估其成熟度（探路负责），仅锁版本 |
| MoeForge/moejs | `v0.1.1` | `b7f49c1f7e502a593d65d4cd9ae3a0b2c3bda070` | 2026-11-16；同上。它对本文唯一的影响点是 §2.2 与 §4 E2 的 goja 替换决策 |
| guoAppGo（新仓库） | 空仓库，零提交 | — | 本文不对它做任何结论 |

核对命令：

```sh
# 原应用（本文依据）
git clone --branch app-v2.0.1-113 --depth 1 https://github.com/Xanerix/guoapp.git
git -C guoapp log -1 --format="%H %ad %s"
git -C guoapp describe --tags     # app-v2.0.1-113

# 两个新栈依赖（本处只锁版本；git ls-remote 对轻量 tag 直接给出 commit）
git ls-remote https://github.com/egoist/mygo.git   "refs/tags/v0.9.2*"   # 25ce62f4…
git ls-remote https://github.com/MoeForge/moejs.git "refs/tags/v0.1.1*"  # b7f49c1f…
```

**tag 与 main 的差异（按要求单独标出，本条已按复核口径更正）**：main 是 `2.0.2+114`（`5a7a116530fcf4242bfc6da08209585e267629cb`，2026-11-17 13:18）。实测 `git diff 4c7aab6 5a7a116`：

| 范围 | 文件数 | 增 / 删 |
| --- | --- | --- |
| 全量 | 83 | +2,058 / −1,032 |
| `lib` + `test`（Dart 应用代码） | **20**（lib 17：15 改 + 2 增；test 3：2 改 + 1 增） | +1,291 / −430 |
| `packages/` | 10 | +493 / −598 |
| `assets/` | 50（49 个 png 为二进制不计行数，1 个文本 LICENSE） | +198 / −0 |
| `native/` | **0** | 0 |

```sh
git -C guoapp fetch --depth 1 origin 5a7a116530fcf4242bfc6da08209585e267629cb
git -C guoapp diff --shortstat 4c7aab6 5a7a116
git -C guoapp diff --name-status 4c7aab6 5a7a116 -- lib test
```

**更正**：本文初稿写"main 比 tag 只多两个文件"，是错的——那次只比对了文件名列表，只能看出新增/删除，漏掉了被修改的 15 个 `lib/` 文件。被改的是：`core_bridge.dart`、`danmaku_controller.dart`、`danmaku_models.dart`、`playback_preferences.dart`、`playback_session.dart`、`ui/detail/detail_page.dart`、`ui/player/danmaku_overlay.dart`、`ui/player/danmaku_settings.dart`、`ui/player/landscape_player_page.dart`、`ui/player/playback_sheets.dart`、`ui/player/player_lock.dart`、`ui/player/player_surface.dart`、`ui/player/portrait_player_page.dart`、`ui/settings/settings_sections.dart`、`ui/widgets/cover_image.dart`。

（行数口径注：按文件逐个数行得到 lib 合计是 tag 37,219 / main 37,902，与 git 的 `+1,153/−430` 方向一致但差约 40 行，原因是 PowerShell 对末行无换行的文件少计一行。引用 tag↔main 差异时以 git 的 numstat 为准。）

**这条差异怎么用——结论比初稿收紧，不是放松**：

1. `native/` 完全一致（0 差异）→ 业务核心在 tag 已定型，§2 的五条前提与 §4 全部优先级**不受影响**。
2. 18 处 Dart 改动**集中在播放线**（逐文件 numstat 见下），且修复与新功能混在一起：播放器锁（#157，新增）、预开下一集改为临近片尾（#159，改）、拖动进度条 5 秒未跳转则重开当前线路（改）、弹幕表情（新增）、macOS 文件选择器 entitlement（改）。
3. **方向提醒：按 main 估不会算多，会算少。** `ui/detail/detail_page.dart` 是 **+97/−315（净删 218 行）**，因为 main 把内容拆进了 `season_strip.dart`（新增 297 行）。只看总数会以为"详情页变小了"，实际是搬走。基线取 tag 才正确。
4. 基线是 tag，意味着 v1 **不继承 2.0.2 的新功能**（弹幕表情、季播条），但**必须把 2.0.2 的修复当作"已知缺陷清单"逐条读**——否则会在取流/重连、播放器锁、进度条跳转上重新踩一遍同样的坑。初稿"不影响判定"的说法把这条风险说轻了。

播放线逐文件 numstat（`git diff --numstat 4c7aab6 5a7a116`，只列净变动绝对值 >10 的）：

| 文件 | 增 / 删 | 净 |
| --- | --- | --- |
| `lib/ui/player/portrait_player_page.dart` | +190 / −24 | +166 |
| `lib/ui/detail/detail_page.dart` | +97 / −315 | **−218** |
| `lib/playback_session.dart` | +54 / −14 | +40 |
| `lib/ui/player/danmaku_overlay.dart` | +78 / −40 | +38 |
| `lib/ui/player/player_lock.dart` | +23 / −17 | +6 |
| `lib/ui/player/player_surface.dart` | +2 / −2 | 0 |

**Windows 渲染的一条旁证**：main 把 `packages/media_kit_video/windows` 的渲染后端从 ANGLE 换成 D3D11（删 `angle_surface_manager.*`，增 `d3d11_surface_manager.*` + `render_d3d11.h`，10 文件 +493/−598）。我们不用 media_kit，不影响实现；但它说明上游在 Windows 渲染上有已知问题正在打补丁——**做 Windows 附加端时 D3D11 这条路是被踩过的，别自己再摸一遍 ANGLE。**

---

## 1. 可复现的规模数字

| 指标 | 数值 | 命令 |
| --- | --- | --- |
| Dart（`lib/`） | 132 文件 / **37,219 行** | `Get-ChildItem -Recurse lib -Filter *.dart \| Get-Content \| Measure-Object -Line` |
| Go（`native/`） | 87 文件 / **13,116 行**（其中 10 个 `_test.go` 共 1,332 行） | 同上，`native` + `*.go` |
| Dart 测试（`test/`） | 28 文件 / 3,600 行 | 同上，`test` |
| 运行时依赖（非 SDK） | **18 个**（见 §8） | `pubspec.yaml` |
| 原生核心 RPC op | **47 个** | `Select-String native/core/app_runtime.go -Pattern '^\s+case "'` |
| 顶层页面文件 | **19 个** `*page*.dart` | `Get-ChildItem -Recurse lib/ui -Filter *page*.dart` |
| 设置分组 | **9 组**（外观/播放/弹幕/下载/存储/代理/用户/关于/诊断） | `lib/ui/settings/settings_sections.dart:29-38` |

### 1.1 子系统体量（决定优先级与工作量归属）

| 子系统 | 位置 | 行数 |
| --- | --- | --- |
| 播放器 UI | `lib/ui/player/`（19 文件） | 8,891 |
| 设置 | `lib/ui/settings/`（9 文件） | 2,966 |
| 详情页 | `lib/ui/detail/`（3 文件） | 2,862 |
| 下载 | `lib/ui/downloads/`（5 文件） | 2,468 |
| 通用组件 | `lib/ui/widgets/`（20 文件） | 3,746 |
| 外壳与主题 | `lib/ui/shell/` + `lib/ui/theme/` | 1,089 |
| 根目录状态/桥接 | `lib/*.dart`（`core_bridge` 1,143 + `local_store` 1,236 + `playback_session` 1,397 为三大头） | 9,397 |
| Go 业务 op 层 | `native/core/app_*.go`（40 文件） | 6,381 |
| Go 站源模块运行时 | `native/core/module_*.go`（11 文件） | 2,614 |
| Go 站源适配层 | `native/core/provider_*.go`（7 文件） | 1,545 |
| Go 网络 | `native/core/network_*.go` + `proxy_*` | 1,344 |
| Go 下载/流/播放 | `app_download*` + `app_stream*` + `app_playback*` | 2,307 |
| Go 投屏 | `app_cast*.go` | 601 |

---

## 2. 五个改变优先级判断的前提

这五条比功能清单本身更重要，先说清楚，否则下面的 P 级会被误读。

**2.1 业务逻辑已经在 Go 里，不在 Dart 里。**
`native/` 是独立的 Go module（`module duanjuapp/native`，`go 1.26.0`），通过一个 C ABI 暴露给 Flutter：`DuanjuRequest(json) -> json` + `DuanjuFree`（`lib/core_bridge.dart:26-68`）。47 个 op 覆盖目录、详情、取流、下载、弹幕、投屏、模块、存储、更新（`native/core/app_runtime.go:464` 起的 dispatch）。下载目录规划、文件命名、Windows 保留名处理也在 Go 里（`app_download_convert.go:21-60`）。
**推论：这不是"把 Flutter 翻译成 Go"。核心业务已有 13k 行 Go 现成代码，重写缺的是 UI 层和平台壳。** 任何按"54k 行全量重写"给的工期都会高估；反过来，把 Go 层当成能零成本复用的也会低估——它现在是"被 FFI 调用的库"，不是"应用主程序"。

**2.2 站源已经是运行时可插拔的 JS 模块。**
模块 = `manifest.json` + `main.js` 的 zip，CommonJS 写法，`module.exports` 导出 `categories/catalogPage/search/detail/resolveMedia/...`，运行时 `github.com/dop251/goja`（`native/core/module_script.go:14,23-33`；示例 `native/module_example/jsexample/`）。manifest 声明 `runtime:"js"`、`api:1`、`capabilities`、`hosts` 域名白名单、`adult`、`allowEmptyCatalog`。模块有实例池（上限 4）、60s 超时、5s 快取超时（`module_script.go:20-22`）。模块商店索引是公开 GitHub JSON（`native/core/module_index.go:25-26`，raw + jsDelivr 双源）。
**决策（2026-11-17，Chief 拍板）：v1 只做红果单源；但必须做模块宿主最小集（M6-lite），moejs 不进主干，ADR-0009 不写。**
修正一处此前的误判：本文初稿把「单源」推成「模块系统整层不做」，这是错的。**红果本身就是 api:1 的 JS 模块**（`manifest.json` + `main.js`，公开仓 `guoapp-modules/hongguo`），仓库里 `native/core/bundled/*` 被 `.gitignore` 排除、只留 `.gitkeep`，构建期由 CI 把红果仓 checkout 进该目录，再由 `//go:embed all:bundled` 编进 Go 核心（`native/core/module_bundled.go`）。**没有模块宿主，红果加载不出来，应用等于白屏。**
所以 v1 的 M6-lite = 构建期注入 + 加载 + 校验 + 复用 goja 执行；**不做**的是商店索引、多源切换、模块安装 UI。省下的是那一层，不是整个模块机制（探路核算为 +5~7 人天，见 `evaluation-plan.md`）。
**推论：`moejs` 的问题不是"要不要引入 JS 引擎"，而是"要不要用 moejs 替掉 goja"。** 这是纯替换题，不是新增能力题，评估口径应该按"迁移成本 + API 兼容性"算，而不是按"引入新依赖的风险"算。

顺带给出隔离层的落点（供 v2 的 moejs ADR 用，不代替探路的结论）：现在跑一段站源 JS 的入口只有两处——`module_script.go` 的 `jsProvider`（建 VM、加载 CommonJS 包装、取 `module.exports`、调用时 JSON 编解码）和 `module_host.go`（给 JS 用的 `host` 对象：`fetch`、`exclusive`、`dramaId`/`chapterId` 等）。**`Provider` 接口以下（`provider_*.go` 1,545 行）与 JS 引擎无关**，已经天然被隔离。也就是说替换 goja 的爆炸半径是这两个文件（475 + 406 行），不是整个 2,614 行的 `module_*`。这一点直接回答“能不能用一层自己的适配层隔离”：能，而且现在就已经隔着，代价是改写约 880 行、并验证 host 对象的语义等价（`fetch` 的超时/限流/重定向/状态码白名单最容易被新引擎的差异咬到）。

**2.3 转码/合成在 Dart 侧，不在 Go。**
`ffmpeg_kit_flutter_new_min_gpl 2.6.7` 承担：下载完成后 mp4 转封装、m3u8 合并、probe 时长、封面解码（`lib/media_pipeline.dart:41-123`、`lib/download_converter.dart:81-148`、`lib/cover_decoder.dart:12`）。Go 侧只做 HLS 抓取与 CENC 解密（`app_download_hls.go`、`media_cenc.go`）。
**推论：Go 重写必须新增 ffmpeg 依赖**（随包分发 CLI，或引入 Go 侧 remux 库），这是新增的外部二进制与签名/公证负担，不是"Go 里已经有了"。

**2.4 播放器完全不在 Go 里。**
`media_kit 1.2.6`（libmpv）+ `video_player`，本仓库还 override 了 `media_kit_video`、`media_kit_libs_windows_video` 两个包（`pubspec.yaml:38-46`）。解码器选择、硬解/软解、清晰度轨切换都在 Dart 侧编排（`lib/playback_session.dart:620-624`）。
**推论：这是整个重写最大的单点不确定。** Go 没有等价播放器，必须自己决定 libmpv(cgo)/webview video/外部进程，每种都影响 UI 形态与性能上限。这一项不定，播放器相关功能（下文 C 组，8.9k 行 Dart）的优先级都只是纸面。

**2.5 原应用有 4 套 UI 形态，新应用只需要 1 套。**
`FormFactor` 有 手机/平板/窗口/电视 四态，自动识别 + 手动锁定（`lib/app.dart:174-184`）；电视模式强制横屏、焦点环、遥控器按键（`lib/ui/widgets/focus_ring.dart` 543 行、`television_player_page.dart` 1,074 行）；竖屏短剧沉浸式（`portrait_player_page.dart` 1,326 行）；手机手势与锁定（`player_lock.dart`、`swipe_controls.dart`）。
**推论：C 组里手机与电视两套播放器是纯增量成本，桌面端只继承"桌面播放器"这一套。** 这块的砍量在原应用里占比最大。

---

## 3. 优先级定义

- **P0**：缺了产品不成立（看完一集剧的闭环）。P0 不做完，v1 不能发。
- **P1**：缺了难用、但能看。发 v1 后补，或 v1 带明显妥协。
- **P2**：锦上添花、长尾、或只对特定端有意义。v1 可以完全不出现。

规模标签按行数与耦合度给，供探路换算人天（不在这里给绝对人天，避免和他的估算打架）：**S** ≤1k 行或纯壳；**M** 1k–3k；**L** 3k–8k；**XL** >8k 或存在未解决的技术选型。

---

## 4. 功能全景（按 25 个功能块）

先给总表，判定依据在 §4.1 起。**「桌面相关性」一列区分：必做 / 可做 / 砍。**

### A. 启动与外壳

| ID | 功能块 | 证据 | 桌面相关性 | 级别 | 规模 |
| --- | --- | --- | --- | --- | --- |
| A1 | 首次启动条款页 + 启动引导 | `ui/terms_page.dart` 182；`ui/app.dart:297` `_BootScreen` | 必做 | P0 | S |
| A2 | 界面形态识别与切换（手机/平板/窗口/电视） | `app.dart:174-200`；`ui/app_scope.dart` 42 | 简化为「窗口」单态 | P0（简化后） | S |
| A3 | 导航外壳（底栏 / 可收起侧栏 / 电视顶栏；宽屏子页在侧栏右侧） | `ui/shell/app_shell.dart` 710 | 必做（侧栏） | P0 | M |
| A4 | 窗口管理（恢复尺寸、置顶、单实例、便携数据、Windows 数据迁移） | `ui/window_frame.dart` 54、`window_on_top.dart` 22、`portable_data.dart` 32、`windows_data_migration.dart` 36、`main.dart:31-53` | 必做（原生重写） | P1 | M |
| A5 | 主题与动效（系统取色 / 6 色板 / 深色 / 纯黑 / 电视主题 / 动效降级） | `ui/theme/app_theme.dart` 103、`motion.dart` 261、`system_color.dart` 15 | 必做 | P1 | M |
| A6 | 全局状态页与错误态 | `ui/widgets/status_view.dart` 80 | 必做 | P1 | S |

### B. 浏览

| ID | 功能块 | 证据 | 桌面相关性 | 级别 | 规模 |
| --- | --- | --- | --- | --- | --- |
| B1 | 站源切换与站源管理入口 | `ui/widgets/source_bar.dart` 84、`ui/explore/source_actions.dart` 44、`source_modules.dart` 107 | 必做 | P0 | S |
| B2 | 发现页（分类 Tab：推荐/全部/榜单/题材 + 无限滚动 + 下拉刷新 + 首屏大卡） | `ui/explore/explore_page.dart` 425、`explore_hero.dart` 300、`ui/widgets/category_tabs.dart` 238、`drama_grid.dart` 254、`drama_tile.dart` 394、`catalog_sort.dart` 80 | 必做 | P0 | L |
| B3 | 榜单（榜单列表 + 分页 + 排名角标） | `Go provider_rankings.go` 133、`ranking_models.dart` 58 | 必做（并入 B2 的分类 Tab） | P0 | S |
| B4 | 搜索（当前源搜索、最近搜索、红果输入联想、首字母键盘） | `ui/search/search_page.dart` 321、`search_field.dart` 187、`letter_keyboard.dart` 231、`Go app_suggestions.go` | 必做 | P0 | M |
| B5 | 详情页（模糊封面底、集数/完结/分类/播放量/标签/简介、续播定位、VIP 标记） | `ui/detail/detail_page.dart` 1419、`poster_glow.dart` 234、`Go app_metadata.go`、`app_seasons.go` | 必做 | P0 | XL |
| B6 | 季与区间选集网格（每组 30 集、季切换） | `ui/player/episode_grid.dart` 436、`Go app_seasons.go` | 必做 | P0 | M |
| B7 | 封面加载/缓存/HEIF 解码/封面修复 | `ui/widgets/cover_image.dart` 330、`cover_loader.dart` 48、`cover_decoder.dart` 201、`Go app_covers.go` 334、`image_heif.go` 344、`app_cover_repair.go` 166 | 必做 | P1 | L |
| B8 | 追剧（追/不追、有更新排序、30 分钟后台检查、搜索、按源筛选、多选取消） | `ui/follow/follow_page.dart` 642、`follow_state.dart` 179、`follow_button.dart` 159、`follow_updater.dart` 93 | 必做 | P0 | L |
| B9 | 观看记录（今天/昨天/更早、搜索、按源筛选、多选删除、清空） | `ui/history/history_page.dart` 631、`watch_history.dart` 23、`Go library_progress.go` | 必做 | P0 | M |
| B10 | 多用户与权限（独立追剧/记录/搜索、管理员增删改、可访问站源、是否允许下载、密码、解锁页） | `ui/users/users_view.dart` 254、`user_editor.dart` 423、`password_dialog.dart` 113、`unlock_page.dart` 177、`Go app_data.go` | 可做 | P1 | L |
| B11 | 我的页（用户卡、最近观看、下载概况、设置与模块入口、下载中角标） | `ui/mine/mine_page.dart` 654 | 必做 | P0 | M |

### C. 播放

| ID | 功能块 | 证据 | 桌面相关性 | 级别 | 规模 |
| --- | --- | --- | --- | --- | --- |
| C1 | 取流与会话（resolve/fallback、多线路重试、进度上报、CENC、本地回环 HTTP 代理、局域网流） | `lib/playback_session.dart` 1397、`playback_loader.dart` 59、`playback_recovery.dart` 49、`Go app_playback.go`、`app_stream.go` 426、`app_stream_cenc.go` 192、`app_playback_routes.go` 151 | 必做 | P0 | XL |
| C2 | 桌面播放器（左视频右选集、宽屏、全屏、悬停控制条、快捷键、滚轮音量、右键宽屏） | `ui/player/player_page.dart` 377、`player_surface.dart` 1278、`playback_page.dart` 300、`playback_sheets.dart` 312、`swipe_controls.dart` 250、`player_status.dart` 69 | 必做 | P0 | XL |
| C3 | 竖屏短剧播放器（上下翻页切集、左右滑、长按倍速、锁定、PiP） | `portrait_player_page.dart` 1326、`mini_player.dart` 355、`picture_in_picture.dart` 282、`player_lock.dart` 134 | **砍**（无手机端） | P2 | — |
| C4 | 电视播放器（焦点环、遥控按键、播控浮层、10 列选集） | `television_player_page.dart` 1074、`focus_ring.dart` 543、`landscape_player_page.dart` 1578（部分） | **砍**（无 TV 端） | P2 | — |
| C5 | 播放偏好（倍速、清晰度、弹幕开关、解码方式、手势开关、记忆） | `playback_preferences.dart` 306、`download_preferences.dart` 26、`local_store.dart` 1236（共用） | 必做 | P0 | M |
| C6 | 预加载与预开下一集（临近片尾触发、暂停超 1 分钟释放） | `playback_preloader.dart` 145、`Go app_preload.go` 257 | 可做 | P1 | M |
| C7 | 弹幕（拉取、滚动渲染、样式/字号/透明度/速度/区域、屏蔽词、表情、设置页） | `danmaku_controller.dart` 353、`ui/player/danmaku_overlay.dart` 560、`danmaku_settings.dart` 394、`danmaku_style.dart` 48、`danmaku_models.dart` 79、`Go app_danmaku.go`、`assets/danmaku_emoji/` | 可做（红果在线源才有） | P1 | L |
| C8 | 投屏（DLNA/SSDP 发现、GBK 设备描述解析、局域网流、控制） | `ui/player/cast_page.dart` 414、`Go app_cast.go` 439、`app_stream.go` 的 `lan` 分支 | 可做 | P1 | M |
| C9 | 清晰度/音轨/倍速切换与解码方式切换 | `playback_sheets.dart` 312、`playback_session.dart:620-624` | 必做 | P0 | M |
| C10 | 播放诊断日记（事件日志、查看、复制） | `diary_service.dart` 89、`ui/settings/diary_view.dart` 138、`Go diagnostics.go` 97 | 可做 | P2 | S |

### D. 下载

| ID | 功能块 | 证据 | 桌面相关性 | 级别 | 规模 |
| --- | --- | --- | --- | --- | --- |
| D1 | 下载队列与任务控制（下载中/已下载分栏、单集与全部暂停/继续/重试、并发数、按源分子目录、角标） | `ui/downloads/downloads_page.dart` 1044、`download_actions.dart` 120、`Go app_downloads.go` 543、`app_download_transfer.go` 305、`app_download_collections.go` | 必做 | P0 | XL |
| D2 | 选集与下载选择器（点选/拖选/全选、VIP 集提示、画质） | `download_picker_page.dart` 239、`ui/player/episode_grid.dart` | 必做 | P0 | M |
| D3 | HLS 抓取 / CENC 解密 / 转普通 mp4 | `Go app_download_hls.go` 244、`media_cenc.go` 277、`app_download_convert.go` 225、`lib/download_converter.dart` 166、`media_pipeline.dart` 126、`ffmpeg_kit` | 必做 | P0 | XL |
| D4 | 旧版分集合并与批量合并队列 | `download_merge.dart` 181、`merge_queue.dart` 218、`ui/downloads/merge_actions.dart` 68 | **砍**（兼容历史数据） | P2 | — |
| D5 | 导出与转移（导出后删副本、移动下载目录） | `download_export.dart` 34、`ui/settings/backup_actions.dart`、`Go app_download_transfer.go` | 简化（桌面本来就在本地） | P1 | S |
| D6 | 已下载剧页（按剧聚合、播放、删除、补齐缺集、编辑） | `downloaded_drama_page.dart` 997、`download_collections.dart` 241、`download_overview.dart` 86 | 必做 | P0 | L |
| D7 | 后台下载（Android 前台服务常驻） | `background_downloads.dart` 126、`flutter_foreground_task` | **砍**（Android 专属） | P2 | — |
| D8 | 已下载文件本地播放（不经网络取流） | `Go localPlayback` op、`local_files.dart` 93、`media_pipeline.dart` | 必做 | P0 | M |

### E. 站源模块与账号

| ID | 功能块 | 证据 | 桌面相关性 | 级别 | 规模 |
| --- | --- | --- | --- | --- | --- |
| E1 | 模块宿主最小集 M6-lite（构建期注入 + 加载 + 校验 + 启用；不含安装 UI、不含多源切换） | `Go module_loader.go` 248、`module_bundled.go` 150、`module_capabilities.go` 135、`native/core/bundled/`（`.gitignore` + `.gitkeep`，CI 注入） | 必做（红果本身是模块） | P0 | M |
| E2 | JS 站源运行时（goja、CommonJS 导出、实例池 4、60s/5s 超时、缓存、host 白名单、能力声明） | `Go module_script.go` 475、`module_host.go` 406、`module_example/jsexample/` | 必做（M6-lite 复用，不换 moejs） | P0 | M |
| E3 | 模块商店索引与安装 UI（四页签、在线仓库、安装/更新/卸载、站源开关） | `ui/modules/modules_page.dart` 797、`module_install.dart` 177、`Go module_index.go` 164 | **砍到 v2**（单源不需要商店） | P2 | — |
| E4 | 站源账号与设备注册（青空登录、设备注册、授权、登出） | `ui/settings/sorani_account_view.dart` 527、`sorani_account.dart` 74、`ui/device_registration.dart` 148、`Go app_devices.go` 214、`sourceAccount` op | **砍**（青空是商店模块源，随 E3 一起消失） | P2 | — |

### F. 设置与数据

| ID | 功能块 | 证据 | 桌面相关性 | 级别 | 规模 |
| --- | --- | --- | --- | --- | --- |
| F1 | 设置壳与 9 分组（即时保存、无权限项置灰并注明原因、宽屏双栏） | `ui/settings/settings_page.dart` 559、`settings_sections.dart` 736、`settings_rows.dart` 164、`appearance_settings.dart` 295 | 必做（可减组） | P0 | L |
| F2 | 代理（系统代理/HTTP/HTTPS/SOCKS5、仅本机保存不进备份、DNS、CDN、限速、浏览器指纹） | `ui/settings/settings_sections.dart:454-477`、`Go proxy_system_windows.go` 167、`network_limiter.go` 264、`network_dns.go` 138、`network_cdn.go` 224、`network_browser.go` 211 | 必做（站源可达性依赖） | P0 | L |
| F3 | 存储（下载目录、移动下载、缓存统计、清缓存、删除全部数据） | `ui/settings/storage_view.dart` 257、`Go app_storage.go` 271 | 必做（可减项） | P1 | M |
| F4 | 备份与恢复（用户/追剧/记录/偏好/主题色/站源开关/模块列表；旧版备份兼容；损坏时从备份恢复） | `ui/settings/backup_actions.dart` 123、`data_card.dart` 171、`Go app_persistence.go`、`app_data.go` 392、`local_store.dart` 1236、`local_snapshot.dart` 95 | 必做（数据安全） | P0 | L |
| F5 | 检查更新（启动检查、忽略此版本、不再提醒）+ 匿名使用统计（可关） | `ui/update_prompt.dart` 132、`app_update.dart` 12、`Go app_update.go` 154、`worker/src/index.js`（`/v1/latest`、`/v1/ping`） | 必做 | P1 | S |
| F6 | 关于（版本、构建、链接、开源许可、条款重看） | `settings_sections.dart:505-540` | 必做 | P1 | S |

---

## 5. 逐块判定依据（P0 与主要 P1）

**A1 条款页 —— P0。** 不是法律洁癖：`terms_page.dart` 是启动唯一闸门（同意前不进主界面）。新应用若沿用同一站源与同一分发方式，这份文案得重写（作者/仓库/免责主体都变了），属于必须产物，不是可选项。

**A3 导航壳 —— P0。** `AppTab` 三页签（发现/追剧/我的）+ 每页独立 Navigator + 宽屏时子页在侧栏右侧打开（`app_shell.dart:26-30, 344-410`）。这套结构决定了后面所有页面怎么写，必须最先定。

**B2 发现页 —— P0，L。** 它是唯一入口：分类 Tab 由站源动态提供（推荐/全部/各榜单/各题材混在一条 Tab 条里，`category_tabs.dart` 238 行），分页游标由 Go 侧 `catalog` op 给（`CatalogPage.cursor`）。这是"能不能找到剧"的全部。

**B3 榜单 —— P0，但并入 B2。** 榜单在 UI 上不是独立页面，是发现页分类 Tab 里的一种（`category_tabs` 与 `rankings` op 共用 Tab 条）。单独拆出来做会重复劳动，**建议与 B2 同一张工单**。

**B5 详情页 —— P0，XL。** 1,419 行里有大量非核心视觉（poster_glow 模糊光晕）。核心是：集数与完结状态、选集入口、续播定位、追剧按钮、下载入口、VIP 标记。**这是最容易被低估的一块**——它不是"展示信息"，是"分发到播放/追剧/下载三个动作的中枢"。

**B8 追剧 —— P0。** 关键不在列表，在**后台 30 分钟轮询未完结的追剧并算出"更新了几集"**（`follow_updater.dart:13-15`：启动延迟 30s、间隔 30 分钟、重查 6 小时）。桌面端要靠定时器常驻，不能靠系统后台任务，实现方式与移动端不同。

**C1 取流与会话 —— P0，XL，且是全项目风险最高的一块。** Go 侧已有 `resolve/fallback`、`app_stream.go` 起本地 HTTP 服务（127.0.0.1 + 局域网两份 server，`app_stream.go:54-127`）、CENC 解密。**但 1,397 行的 `playback_session.dart` 才是真正的编排层**：多线路回退、进度上报、错误恢复、轨切换、下一集预开。这段逻辑现在在 Dart 里，重写必须整体搬到 Go 或新 UI 层——**它既不在 Go 里，也不是纯 UI**，是"半个播放器内核"。

**C2 桌面播放器 —— P0，XL。** 取决于 §2.4 的选型。若走 webview + `<video>`，`player_surface.dart` 的手势/快捷键/悬停控制条逻辑可大部分保留思路，但要重写；若走 libmpv(cgo)，渲染与控件要自己搭，工作量与风险都更高。**这一块没有"照抄"选项。**

**C7 弹幕 —— P1 而非 P0。** 只有内置红果在线播放有弹幕，本地播放不加载（README:48）。若 v1 只做桌面 + 红果，弹幕是"体验项"；若不做，红果与其他源的差距不大。表情图（`assets/danmaku_emoji/`）在 tag 里没有对应 Dart 代码（main 才有 `danmaku_emoji.dart`），**tag 基线不含弹幕表情**，注意别按 main 的功能给工作量。

**D1/D3 下载 —— P0，XL，跨两个语言栈。** Go 侧完成抓取/解密/命名，Dart 侧完成转封装。**重写时 D3 的 ffmpeg 部分等于从零补**（§2.3）。原应用"每集下载完自动转成普通 mp4，可直接拷出"是这款产品的核心卖点之一（README:60），不是可选功能。

**E1/E2 模块宿主 —— P0（按 M6-lite 口径，先前的"整层砍"是误判）。**
关键事实：**红果是模块，不是内置代码**。`native/core/bundled/*` 在 `.gitignore` 里（只保留 `.gitkeep`），构建期 CI checkout `guoapp-modules/hongguo`，再由 `//go:embed all:bundled` 嵌进核心（`native/core/module_bundled.go`）。所以 v1 必须要 `module_script.go` + `module_host.go` + `module_loader.go` + `module_bundled.go` 这一套，否则红果加载不出来。
v1 做的 M6-lite = **构建期注入 + 加载 + 校验 + 复用 goja 执行**；v1 不做的是 `module_index.go`（商店索引）、模块安装/更新 UI（`modules_page.dart` 797 + `module_install.dart` 177）、多源切换。
补两条来自 2.0.1 之后的证据，说明模块机制不能只当"M6-lite 小切片"看：
- 分类缓存要按模块版本失效（#169，`619846e`：缓存记录模块版本，版本变化重新拉分类），否则升级后红果分类里找不到听书和小说——**M6-lite 必须带模块版本号**。
- 红果取流问题是通过升模块版本修的（`33b19a7`：只返回 H.266 线路导致取流失效，改用低端机档位，模块 1.0.3）——**站源修复走模块发版，不走应用发版**，这正是模块化的价值，也是 M6-lite 的最小可维护边界。

**E3 在线商店 —— 砍到 v2。** 商店索引是公开 GitHub JSON（`native/core/module_index.go:25-26`，raw + jsDelivr 双源），v1 单源不需要它。**E4 站源账号（青空等）—— 砍**，青空是商店里的模块源，随 E3 一起消失。

**连带影响：C7 弹幕的优先级要往上提。** 单源之后红果是唯一内容来源，而弹幕只有红果在线播放有（README:48），它从"多源里的一个体验项"变成"唯一源的差异化体验"。本清单仍按 P1 记，但 v1 范围评审时请把 C7 和 C2 放一起看，不要按"可有可无"处理。

**F2 代理 —— P0。** 站源可达性直接依赖代理（HTTP/HTTPS/SOCKS5 + 系统代理跟随 + 浏览器指纹伪装，`network_browser.go` 211 行用 `bogdanfinn/tls-client`）。这不是"高级设置"，是没有它部分源直接 403（CHANGELOG `2.0.1+113` 第 19 条就修的是这个）。

**F4 备份 —— P0。** 数据只在本地（无账号云同步），备份是唯一的数据保险；且备份格式要与旧版兼容（README:67 "旧版备份仍可导入"）。重写若不兼容旧备份，老用户迁移会丢追剧与记录。

---

## 6. 可以砍掉的（v1 不做）

| 砍掉的 | 原规模 | 理由 | 代价 |
| --- | --- | --- | --- |
| C3 竖屏短剧播放器 + 竖屏手势 + 解锁 | ~2.1k 行（`portrait_player_page` 1326 + `mini_player` 355 + `player_lock` 134 + 相关分支） | 新应用无手机端 | 无。桌面用户不会想念上下翻页 |
| C4 电视播放器 + 焦点环体系 | ~1.6k 行（`television_player_page` 1074 + `focus_ring` 543） | 新应用无 TV 端 | 无。但注意**电视模式也影响了 `app_shell`、`landscape_player_page`、`detail_page` 的分支**，删功能时要把这些分支一起清掉，否则留死代码 |
| 界面形态四态 + 手动锁定 | `app.dart:174-200` 与散落各处的 `television` 分支 | 桌面只有「窗口」一态 | 无。省下的不只是代码，是**每次改 UI 都要考虑的 4 种布局** |
| D4 旧版分集合并 + 批量合并队列 | ~470 行 | 只服务旧版本地数据（README:61） | 老用户旧下载不自动合并，需手动处理 |
| D7 Android 前台服务后台下载 | 126 行 + 整个 `flutter_foreground_task` 依赖 | Android 专属 | 无 |
| B10 多用户 + 权限 + 密码 | ~970 行 UI + `app_data` 里的用户域 | 桌面单人使用场景弱；且原应用多用户是为"电视/家庭共享"设计的 | 家庭共用一台 Mac 的场景会退化为单用户 |
| E3 在线商店与模块安装 UI（四页签、索引、安装/更新/卸载） | 974 行 UI（`modules_page` 797 + `module_install` 177）+ Go `module_index.go` 164 | 单源不需要商店；红果靠构建期注入 | 装不了第三方源，多源放 v2 |
| E4 站源账号（青空等） | ~750 行 | 青空是商店模块源，随 E3 一起消失 | 无（v1 不需要） |
| C8 投屏 | ~1k 行（414 UI + 439 Go + lan server） | 桌面投屏是长尾；且 DLNA 依赖同一局域网 | 想电视上看要自己拷文件 |
| A4 便携数据 / Windows 数据迁移 | ~120 行 | 历史包袱（从便携版迁到用户目录） | 新应用没有历史数据，天然不需要 |

**砍完这几项，原应用 37,219 行 Dart 里约 8–9k 行直接归零。Go 侧按 M6-lite 只少搬 `module_index.go`（164 行），`module_script/host/loader/bundled`（约 1.4k）仍然要沿用——先前"整块 E 省 2.6k 行 Go"的说法作废，别拿它当省人天的依据。**

---

## 7. 可以简化的（v1 保留功能但降级实现）

| 简化项 | 原来 | 建议 | 省下 |
| --- | --- | --- | --- |
| A5 主题 | 系统取色 + 6 预设色板 + 纯黑 + 电视主题 + `motion.dart` 261 行动效体系 | 系统深/浅色 + 2 个预设色板；不做动效体系，用平台原生过渡 | ~400 行，且少一层全局包装 |
| B7 封面 | 远程加载 + 本地缓存 + HEIF 解码（`image_heif.go` 344）+ 封面修复 op + 模糊底 | 远程 + 磁盘缓存；HEIF 交给系统解码；不做"修复" | ~600 行，少一个 ffmpeg 用途 |
| C7 弹幕 | 滚动渲染 + 完整样式矩阵 + 设置页 394 行 | 只做开关 + 透明度 + 字号；不做区域/速度/屏蔽词 | ~700 行 |
| D5 导出 | 下载完成后自动导出到外部目录 + 删副本 + 手动导出 | 桌面端下载目录本身就是普通目录，"导出"退化为"打开所在文件夹" | 整块可退化为一个按钮 |
| D6 已下载剧页 | 997 行：按剧聚合 + 编辑 + 补齐缺集 + 迁移 | 按剧聚合 + 播放 + 删除；不做补齐/编辑 | ~400 行 |
| F1 设置 | 9 组 736 行 + 权限置灰体系 | 6 组（外观/播放/下载/站源/网络/关于）；权限置灰随多用户一起去掉 | ~500 行 |
| F3 存储 | 目录迁移 + 缓存统计 + 清缓存 + 删全部数据 | 只保留"下载目录"与"删全部数据" | ~200 行 |
| C10 诊断 | 事件日志、查看、复制、上报 | 只写滚动日志文件到本地，不提供 UI | ~230 行 |
| F5 更新 | 启动检查 + 忽略版本 + 使用统计上报 | 只做检查更新；统计可选，甚至不做 | ~100 行 + 一个 worker 依赖 |

---

## 8. 端差异对照：原 4 端 → 新 2 端

| 能力 | Android | iOS | Windows | macOS | 新应用（macOS arm64 主 / Win x64 附带） |
| --- | --- | --- | --- | --- | --- |
| 界面形态 | 手机 | 手机 | 窗口/平板 | 窗口/平板 | 只做窗口 |
| 播放器 | 竖屏沉浸 + 手势 | 竖屏沉浸 + PiP | 桌面布局 + 快捷键 | 桌面布局 + 快捷键 | 桌面布局 |
| PiP | 系统 PiP | 系统 PiP | 自绘小窗（`windows` 分支，`picture_in_picture.dart:157`） | 自绘小窗 | **需确认是否要**（macOS 有原生 PiP，但实现路径不同） |
| 后台下载 | 前台服务 | 受限 | 无（进程在就下） | 无 | 无 |
| 下载转码 | ffmpeg_kit | ffmpeg_kit | ffmpeg_kit | ffmpeg_kit | **需自行捆绑 ffmpeg** |
| 投屏 | DLNA | DLNA | DLNA | DLNA | 同左（Go 侧已实现，可直接复用） |
| 系统取色 | Android 12 壁纸 | — | DWM 强调色 | NSColor 强调色 | 需重写（`system_color.dart` 只有 15 行，成本低） |
| 单实例 | — | — | 有（CHANGELOG #116） | — | 建议 macOS 也加 |
| 便携数据 | — | — | 有 | — | 砍 |

**注意一条隐藏工作量**：`core_bridge.dart` 里每个平台走不同的库加载方式（Android `.so` / Windows `.dll` / iOS+macOS `DynamicLibrary.process()`），而 `DynamicLibrary.process()` 在 macOS 上意味着 **Go 核心被静态链进 App 二进制**。新方案如果把 Go 当主程序（而非被链入的库），这套加载与生命周期管理整体作废、换成 `main()`——**这不是"少写代码"，是换了一种进程模型，得重写数据目录、单实例、退出清理（`release`/`cancelPlayback`/`workLease` 这些 op 现在是为了让 Flutter 侧能优雅关停而存在的）**。

---

## 9. Flutter 依赖 → 重写缺口映射

| 依赖 | 用途 | Go / 新栈现状 | 缺口 |
| --- | --- | --- | --- |
| `media_kit` + `media_kit_video` + `media_kit_libs_video`（含两个本地 override 包） | 播放器（libmpv） | **无** | **最大缺口**，见 §2.4 |
| `video_player` | 备用播放路径 | 无 | 同上 |
| `ffmpeg_kit_flutter_new_min_gpl` | 转封装/合并/probe/封面解码 | **无**（Go 只做抓取与解密） | 需捆绑 ffmpeg CLI 或 Go remux 库；涉及签名、公证、体积 |
| `ffi` | Dart↔Go 桥 | Go 即主程序后不需要 | 无（反而是简化） |
| `window_manager` | 窗口尺寸/置顶/全屏 | 无 | 需 mygo 或平台 API（Cocoa/Win32） |
| `dynamic_color` | 系统取色 | 无 | 平台 API 读取，S |
| `file_picker` | 选 zip 模块 / 选目录 / 恢复备份 | 无 | 平台文件对话框，S |
| `path_provider` + `shared_preferences`（含本地 override） | 数据目录与偏好存储 | Go 侧已有 `app_storage`/`app_data`/`app_persistence` | **可直接复用 Go 侧**，Dart 侧那套作废 |
| `flutter_foreground_task` | Android 后台下载 | — | 砍 |
| `url_launcher` | 打开链接 | 无 | S |
| `crypto` | 哈希（备份校验等） | Go 标准库 | 无 |
| `package_info_plus` | 版本号 | 编译期常量 | 无 |
| `lpinyin` | 拼音排序/首字母键盘（搜索用） | 无 | 只有搜索字母键盘需要，随 B4 简化可砍 |
| `flutter_localizations` | 控件本地化 | 无 | **注意：应用没有 i18n 文案层**，全部中文硬编码（`*.arb` 文件数为 0）。新应用可直接中文写死，不要引入 i18n 框架 |

---

## 10. 给探路的拆分口径（把功能块挂到人天上）

建议按这 7 张工单拆，与 §4 的 ID 一一对应，避免重复计工：

1. **壳与导航**：A1 A2 A3 A4 A5 A6（桌面单态）→ 对应 M+S 若干
2. **浏览闭环**：B1 B2 B3 B4 B5 B6 B11
3. **追剧与记录**：B8 B9（含后台轮询）
4. **播放**：C1 C2 C5 C9（+C6 C7 视取舍；单源后 C7 弹幕权重上调，见 §5）
5. **下载**：D1 D2 D3 D6 D8（+D5 简化版）
6. **模块宿主与网络**：E1 E2（M6-lite：注入 + 加载 + 校验 + goja）+ F2 代理（E3 商店与 E4 账号不做，放 v2）
7. **设置与数据**：F1 F3 F4 F5 F6 A5

**Go 核心复用不计入以上任何一张**：`native/` 13,116 行里，`app_*`（6,381）与 `provider_*`（1,545）基本可直接沿用，`app_stream`/`app_cast`/`media_cenc`（约 1.4k）可沿用；`module_*`（2,614）按 M6-lite 搬 `script`/`host`/`loader`/`bundled`（约 1.4k），**不搬** `module_index.go`（164）与安装 UI 那套。**但"沿用"不等于零成本**：现在的 Go 是库，要改成主程序并重新定义生命周期与并发模型，这笔"改造费"应单列一张工单，不要摊进功能里。

---

## 11. 需要产品确认的问题（按阻塞程度排）

**已定（2026-11-17，Chief 拍板，本群讨论）：v1 只做红果单源；模块宿主最小集 M6-lite 保留**（构建期注入 + 加载 + 校验，复用 goja），不做商店索引与多源切换，moejs 不进主干，ADR-0009 不写。原问题留档如下，不再待决。

1. ~~**v1 到底要不要多站源？**~~ → **已定：单源。** 代价是站源成单点、少了"多源"这个差异点；多源放 v2 复评（见 §2.2、§5 E 段）。连带影响：C7 弹幕从"体验项"上升为唯一源的差异化体验，v1 范围评审时与 C2 一起看。
2. **播放器路线定没定？**（阻塞 C1/C2 全部估算）libmpv(cgo) / webview video / 外部进程，三条路的工期差 2 倍以上。**现状：等 R0（播放能否塞进 webview）的 spike 结论，暂不锁定**，判定门槛与失败路径见 `evaluation-plan.md` 第 7 节。给 R0 补一条方法要求：**必须拿原 Flutter 应用作对照组**（同源、同网络、同 30 分钟），否则 R0 失败时分不清是"webview 撑不住"还是"tag 基线自带的取流缺陷被搬过来了"（§0 第 4 点：2.0.2 修的正是取流/重连），会把架构问题误判成实现问题。
3. **ffmpeg 允许随包分发吗？**（阻塞 D3）涉及安装包体积（几十 MB）、macOS 公证、GPL 传染性（原项目本身是 GPL-3.0，新项目许可证未定）。
4. **要不要兼容旧版备份？**（阻塞 F4）不兼容 = 老用户丢追剧与记录；兼容 = 要读懂旧格式。
5. **多用户要不要？**（阻塞 B10 ≈ 970 行 UI）桌面单人场景弱，但原应用按家庭共享设计的。
6. **弹幕要不要？**（阻塞 C7）只有红果在线有；tag 基线**不含**弹幕表情（那是 main 的 `2.0.2`）。
7. **macOS 要不要系统原生 PiP（画中画）？**（阻塞 C3 的替代物）原应用 Windows 是自绘小窗，macOS 走哪条没定。
8. **条款文案谁出？**（阻塞 A1）主体、仓库、免责范围都变了，不能照抄。

---

## 12. 查不到 / 本文未覆盖

- **原应用的实际用户量与使用分布**：查不到（仓库无公开统计，`worker` 只发匿名计数且无公开数据）。这影响"砍多用户/投屏"的风险评估，但无法量化。
- **`media_kit` 两个本地 override 包改了什么**：`packages/media_kit_video`、`packages/media_kit_libs_windows_video` 在仓库内，本次未逐个 diff；若走 libmpv 路线需要细看，属探路/戈的范围。
- **红果取流的签名算法（v5 签名，CHANGELOG #165）**：在 `provider_*` 与模块内，是"能不能播"的命门，本次只确认了存在与位置，未验证可移植性。**建议戈优先验这条**。
- **`guoapp-sources` 私有仓库**：`CLAUDE.md` 提到官方站源正从 Go 改写成 JS，仓库不可访问，无法统计其规模。
- **`worker/` 线上部署状态**：代码在仓库里（Cloudflare Worker，两个路由），线上是否仍在服务未验证。
- **2.0.2 的修复明细（本文最大遗漏）**：基线是 tag `app-v2.0.1-113`，`2.0.2+114` 上 15 个 `lib/` 文件、`+1,291/−430` 的改动只核到了文件和行数，**没逐条核过每条改动在修什么缺陷**。须对照 `CHANGELOG.md` 的 `2.0.2+114` 一节逐条过，否则 v1 会把已修的坑重新踩一遍（已确认涉及：取流/重连、播放器锁、拖动进度条跳转、弹幕表情、macOS 文件选择器 entitlement）。
