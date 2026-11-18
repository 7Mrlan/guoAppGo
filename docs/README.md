# guoAppGo 文档索引

本目录是 guoAppGo 的唯一文档入口。仓库：https://github.com/7Mrlan/guoAppGo

## 目录结构

```
docs/
  README.md                                本文件：索引 + 文档纪律 + 产出填写要求
  versions.md                              版本基准：原应用 main 提交 + mygo/moejs 钉死 tag，唯一事实来源
  architecture.md                          架构：模块划分、进程/线程模型、IPC、数据流
  platform.md                              macOS arm64 与 Windows 11 x64 的差异清单
  evaluation.md                            评估一：0 版本基准 → 1 结论 → 2 规模 → 3 MyGo → 4 moejs → 附录 A
  evaluation-plan.md                       评估二：5 路线 → 6 模块清单 → 7 风险与验证 → 8 待验证 → 附录 B
  feature-inventory.md                     功能块清单与 P0/P1/P2 分级
  spike-stack.md                           可行性验证结果与踩坑
  plan-m2.md                               当前阶段（M2）任务清单：基准提交、任务编号、可执行验收
  adr/README.md                            ADR 编号规则、模板入口、标题与状态清单
  adr/0000-template.md                     ADR 模板
  adr/0002-ui-route-webview.md             已接受（条件：R1 B 段不劣于对照组）
  adr/0003-playback-kernel-hlsjs-proxy.md  已接受（hls.js + 本地代理，暂不嵌 mpv）
  adr/0006-frontend-vue-vite.md            已接受（Vue 3 + TS + Vite，产物进库）
  testing/acceptance.md                    功能验收标准的写法与清单
  testing/go-test-plan.md                  go test 用例清单与覆盖率口径
  testing/regression.md                    发版前回归清单（含结果记录）
```

第一轮评估产出（不属于上面固定骨架，属于交付物）。**入口固定为 [evaluation.md](evaluation.md)，按下列顺序读**：

1. [evaluation.md](evaluation.md) — 探路（一）：0 版本基准 → 1 结论 → 2 原始规模 → 3 MyGo → 4 moejs → 附录 A 复现命令
2. [evaluation-plan.md](evaluation-plan.md) — 探路（二）：5 路线取舍 → 6 工作量估算 → 7 风险与验证 → 8 查不到/待验证 → 附录 B
3. [feature-inventory.md](feature-inventory.md) — 理清：功能块清单与优先级
4. [spike-stack.md](spike-stack.md) — 戈：0 版本基准 → 1 环境阻塞 → 2 MyGo → 3 moejs → 4 交叉编译 → 5 MediaSource → 6 MSE 实测 → 7 坑 → 8 结论 → 9 未覆盖

评估两份是同一份评估按章节边界切开，不是两份独立报告：章节编号连续（0–4 接 5–8，不重起），两文互相链接，任一处链接失效即视为断链。`evaluation-plan.md` 的数据与结论一律引用 `evaluation.md`，不得各自另立一套数字。

## 当前阶段

M2 进行中，清单见 [plan-m2.md](plan-m2.md)：范围以基准提交 `012de5f6`（`012de5f6ed67969ca57df03ccca1225018427277`）为界，开工前先对基准——`git fetch origin main && git rev-parse origin/main` 与文档比对，并用 `git rev-parse --is-shallow-repository` 排除浅克隆导致的历史截断（曾因此把 34 个提交数成 401）。任务编号、验收与顺序以该文件为准，完工记录回填到 [testing/acceptance.md](testing/acceptance.md)。

## 验证脚手架

可行性验证的可执行部分在 [`tools/spike/`](../tools/spike/README.md)，三个探针分别对应 [spike-stack.md](spike-stack.md) §3、§5、§6。不属于产品代码，不改动时不随产品提交。

| 探针 | 对应章节 | 验证内容 |
| --- | --- | --- |
| `moejs-eval` | §3 | moejs v0.1.1 嵌入 Go、调用 JS、异常与超时 |
| `webview-features` | §5 | 系统 webview 的 MediaSource / ManagedMediaSource / 编解码支持 |
| `mse-chain` | §6 | 本地 fMP4 HLS → hls.js → MSE，首帧与 seek 耗时 |

三个探针的依赖版本钉死为 mygo `v0.9.2`（`25ce62f`）、moejs `v0.1.1`（`b7f49c1`），与 [versions.md](versions.md) 一致；升级须先改 versions.md 再改探针。

### macOS 真机一页单

[MAC-RUN.md](../tools/spike/MAC-RUN.md)：在个人 Mac 上用 10 分钟跑完三项验证的操作单——三个探针的本机编译方式、`go tool mygo dev` 与裸二进制的取舍、`features.json` 与 `mse-chain.json` 的回传方式、字段判读与通过线（首帧 ≤2000 ms、seek ≤1500 ms、内存按私有内存计）。

它销的是 [spike-stack.md](spike-stack.md) §5.2 与 §9 的 macOS 挂账。产出未回传前，凡涉及 macOS 的项按下方纪律第 7 条写 **未验证**。

## 文档自检

```
node scripts/check-docs.mjs
```

无外部依赖，需 Node 18+。退出码 0 全过、1 有失败，失败项逐条打印 `文件:行号` 与原因。

| 检查项 | 规则 |
| --- | --- |
| 链接与文件形态 | 所有 md 的相对链接目标必须存在；文件必须 UTF-8、无 BOM、纯 LF、末尾有换行 |
| 章节编号连续性 | [evaluation.md](evaluation.md) 为 0–4 + 附录 A，[evaluation-plan.md](evaluation-plan.md) 为 5–8 + 附录 B，编号不得跳号或重起 |
| 版本基准确性与一致 | [versions.md](versions.md) 的 guoapp 行必须写 main 提交号并标明以 main 为准、必须保留 `app-v2.0.1-113` 阅读锚点；mygo/moejs 必须仍为钉死 tag；引用 40 位 sha 的文档必须有"## 0. 版本基准"节、必须链接 versions.md、且 sha 不得出自基准之外；`go.mod` 与 `package-lock.json` 的 mygo 版本必须一致且都已登记 |
| README 登记 | 目录树与链接里登记的文件必须存在；`docs/` 下每个 md 必须在本文登记 |

改文档、增删文件、调整章节编号后必须跑一遍；检查失败不得提交。

## 文档纪律

1. 文档只写事实、决定、可复核的引用。不写"我们计划""未来可能"。
2. 每个数字带来源：命令输出、文件路径、版本 tag 或 commit sha、issue 链接。查不到就写"查不到"，并写出已尝试的检索方式。
3. 不确定的结论标 **待验证**，并写出验证方式。
4. 代码行为变化与文档修改在同一次提交内完成。提交权在把关。
5. 版本号、路径、依赖名必须可查，不得凭印象填写。
6. 版本一律按 tag 锁定并附 commit sha，来源见 [versions.md](versions.md)。禁止 `@latest`、禁止按 main 分支下结论。
7. macOS 未拿到真机前，凡涉及 WKWebView 实际行为、`.app` 打包、签名、公证的项一律写 **未验证**，不得写"已通过"；Windows 实测结论不得外推为两端结论。
8. 文档不写工作量或时间估算，也不写进度预测。进度只以"该阶段验收标准是否全部通过"衡量。

## 第一轮产出的填写要求

### docs/evaluation.md（负责人：探路，评估一：0–4 节）

第 0 节为版本基准，其余与原评估任务的前三点对应：

0. 版本基准：原应用 `app-v2.0.1-113`、mygo `v0.9.2`、moejs `v0.1.1`，各自写 tag + commit sha（取自 [versions.md](versions.md)），并声明后续结论以此为准
1. 结论：先给可执行结论（建议做/不做、推荐哪条路线、量级），细节放后面
2. 原始规模：代码行数、页面数、状态管理方案、依赖清单
3. MyGo 现状：版本、原生 UI 与 webview 两条路线成熟度、平台能力矩阵、已知坑
4. moejs 现状：版本、API 稳定性证据（release 与 breaking 记录）、维护活跃度、停更后的代价与隔离手段
附录 A：复现上述数字的完整命令

### docs/evaluation-plan.md（负责人：探路，评估二：5–8 节）

章节编号接续上一份，不重起：

5. 路线取舍：全 Go 原生 UI / Go+webview / 混合三种形态，每条必须写否决条件
6. 模块清单：模块职责、依赖关系、复用来源与取舍理由；**不写工作量或时间估算**
7. 风险与验证：前 5 风险及每个风险的验证方式，一条命令或一个半天内能做完的 spike
8. 查不到 / 待验证：显式列出没查到依据的项，并写已尝试的检索方式
附录 B：给下一轮的输入

硬性要求（两份评估共用）：

- 规模数据从仓库实拉，附命令与输出（例如 `git ls-files | wc -l`、`cloc` 结果、`pubspec.yaml` 原文）。
- 依赖与版本写具体 tag 或 commit sha，不写"最新"。
- 不出现任何工作量或时间估算——这是硬性禁止项，不是风格建议；结论只靠技术依据与验收结果。
- 现状类结论必须带证据（release 记录、issue 链接、源码行数与路径）；"查不到"要写已尝试的检索方式。
- 禁止无来源数字、禁止"大概/可能"。

### docs/feature-inventory.md（负责人：理清）

功能块表格列（缺列即为未完成）：

| 编号 | 功能块 | 原实现位置 | 用户可见行为 | 优先级 | 判定依据 | 重写处置 | 依赖块 |

- 优先级仅取 P0/P1/P2，判定依据写清"为什么是 P0"（例如：去掉后无法播放内容）。
- "原实现位置"必须是 `app-v2.0.1-113` 这个 tag 下的真实路径，不能凭印象写文件名。
- 重写处置仅取 保留 / 简化 / 砍掉，砍掉项必须写理由。
- 文末必须有：P0 最小可发布集清单、砍掉项汇总。

### docs/spike-stack.md（负责人：戈）

每条验证一个二级标题，必含字段：

- 验证目标
- 环境：操作系统版本、Go 版本、mygo 版本（tag + sha）、moejs 版本（tag + sha）
- 代码来源：按 tag 拉取的命令，与 [versions.md](versions.md) 一致
- 执行的完整命令（可复制）
- 原始输出（关键片段，含耗时；失败必须贴错误原文）
- 结论：通过 / 失败 / 部分通过
- 踩坑与绕过方式

交叉编译结论必须区分"编译通过"与"产物可运行"，两者不能合并成一句"能编译"。
