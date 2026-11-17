# guoAppGo 文档索引

本目录是 guoAppGo 的唯一文档入口。仓库：https://github.com/7Mrlan/guoAppGo

## 目录结构

```
docs/
  README.md              本文件：索引 + 文档规范 + 三份第一轮产出的填写要求
  versions.md            版本基准：三个依赖的 tag + commit sha，唯一事实来源
  architecture.md        架构：模块划分、进程/线程模型、IPC、数据流
  platform.md            macOS arm64 与 Windows 11 x64 的差异清单
  adr/                   ADR 目录：编号规则、模板、标题清单
  testing/
    acceptance.md        功能验收标准的写法与清单
    go-test-plan.md      go test 用例清单与覆盖率口径
    regression.md        发版前回归清单（含结果记录）
```

第一轮评估产出（不属于上面固定骨架，属于交付物）：

```
docs/evaluation.md          探路：规模/栈现状/路线取舍/工作量/风险
docs/feature-inventory.md   理清：功能块清单与优先级
docs/spike-stack.md         戈：技术栈最小验证的命令与结果
```

## 文档纪律

1. 文档只写事实、决定、可复核的引用。不写"我们计划""未来可能"。
2. 每个数字带来源：命令输出、文件路径、版本 tag 或 commit sha、issue 链接。查不到就写"查不到"，并写出已尝试的检索方式。
3. 不确定的结论标 **待验证**，并写出验证方式。
4. 代码行为变化与文档修改在同一次提交内完成。提交权在把关。
5. 版本号、路径、依赖名必须可查，不得凭印象填写。
6. 版本一律按 tag 锁定并附 commit sha，来源见 [versions.md](versions.md)。禁止 `@latest`、禁止按 main 分支下结论。

## 三份第一轮产出的填写要求

### docs/evaluation.md（负责人：探路）

必填章节。第 0 节为版本基准，其余与评估任务的六点一一对应：

0. 版本基准：原应用 `app-v2.0.1-113`、mygo `v0.9.2`、moejs `v0.1.1`，各自写 tag + commit sha（取自 [versions.md](versions.md)），并声明后续结论以此为准
1. 原始规模：代码行数、页面数、状态管理方案、依赖清单
2. MyGo 现状：版本、原生 UI 与 webview 两条路线成熟度、平台能力矩阵、已知坑
3. moejs 现状：版本、API 形态、性能口径、是否值得承载插件/脚本
4. 三条路线取舍：全 Go 原生 UI / Go+webview / Go+moejs 混合
5. 工作量估算：按模块拆，乐观/现实/悲观三档，单位人天
6. 前 5 风险及每个风险的验证方式

硬性要求：

- 规模数据从仓库实拉，附命令与输出（例如 `git ls-files | wc -l`、`cloc` 结果、`pubspec.yaml` 原文）。
- 依赖与版本写具体 tag 或 commit sha，不写"最新"。
- 估算表必须写清口径：几个人、是否熟悉 Go、是否包含测试与文档工时。
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
