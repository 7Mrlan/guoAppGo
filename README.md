# guoAppGo

小果影视（guoapp）的 Go 桌面重写版。

技术栈：MyGo（桌面框架）+ moejs（JS 运行时）。
目标平台：macOS arm64 为主，Windows 11 x64 附加。

路线：**B——Go 核心 + webview 前端**（列表与播放页都在 webview 内，播放走 hls.js + 本地代理，暂不嵌 mpv）。按 **117 人天**立项（v1 单源含 M6-lite 模块宿主 115~117，对外报保守值；单源不含模块宿主为 99 小计 / 109 合计，明细见 [docs/evaluation-plan.md](docs/evaluation-plan.md) §6.3）。这是**带条件的决定**：真源实测（`tools/spike/MAC-RUN.md` 同口径的 B 段）不劣于原 Flutter 对照组才确认，不达标则走失败路径：单源合计 **136~150 人天**，高于 A 单源的 133，按"**失败即选 A**"处理（见 [docs/evaluation-plan.md](docs/evaluation-plan.md) §6.3 条件说明）。

平台现状：团队没有 Mac 真机，v1 验收以 Windows 为准；macOS 的 WKWebView 行为、`.app`、签名、公证一律**未验证**，条件编译保留但不进 v1 验收。

## 版本基准

| 项目 | 版本 |
| --- | --- |
| Xanerix/guoapp（原应用） | main 提交 `012de5f6`（`012de5f6ed67969ca57df03ccca1225018427277`，2026-11-17T16:54:50Z） |
| 阅读锚点 tag | `app-v2.0.1-113`（`4c7aab6`）：评估数据在此 tag 上实测，只作阅读锚点 |
| egoist/mygo | `v0.9.2`（钉死，不跟 main） |
| MoeForge/moejs | `v0.1.1`（钉死，不跟 main） |

原应用以 **main 最新提交**为基准，不以 tag 收工：每次开工前先取新增提交，把上次基准之后的功能并进待办。MyGo 与 moejs 的版本钉死不动。

完整 commit sha、来源与锁版规则见 [docs/versions.md](docs/versions.md)。该文件是全项目唯一的版本事实来源。

## 构建与运行

需要 Go 1.27.4（与 `go.mod` 一致）。MyGo 锁在 `v0.9.2`，前端在 `web/dist` 内由 `go:embed` 打进二进制，不需要 Bun 或 Vite。

```sh
go build -o guoappgo .    # 产物在 .gitignore 之外，自行指定输出名
./guoappgo                # 打开一个窗口，页面由二进制内嵌提供
go test ./...             # 不需要显示器，测窗口选项、版本与内嵌前端
```

`.app` / `.dmg` 打包（`mygo build`）与签名、公证属于 macOS 侧工作，M0 不做，也没在真机验证过。

## 文档

文档入口：[docs/README.md](docs/README.md)
