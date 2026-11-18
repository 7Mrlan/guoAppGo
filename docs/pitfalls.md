# 踩坑登记

每条固定四段：**现象 → 根因 → 正确做法 → 怎么验证**。只登记已经踩过并有结论的坑，不写"注意一下"这类提醒。

适用：本仓库的文档、构建、测试、spike 与多人协作。新坑当场补一条，**与修复同一次提交**。

---

## 1. 浅克隆下统计提交距离

**现象**：`git rev-list --count 4c7aab6..012de5f6` 报出 401，与 GitHub compare 的 34 相差一个量级，据此排范围会把清单放大十几倍。

**根因**：之前执行过 `git fetch --depth 1 origin refs/tags/...`，仓库变成浅克隆，父链被截断，`rev-list` 在截断后的历史上算出差值。`--filter=blob:none` 的完整克隆同样会中招——只要再叠一次 `--depth 1`。

**正确做法**：数任何提交距离前，先判是否浅克隆；是就先补全。

```bash
git rev-parse --is-shallow-repository   # 必须是 false 才继续
git fetch --unshallow                   # 为 true 时先补全
git rev-list --count <base>..<head>
```

**怎么验证**：本地数与远端对照，两者一致才算数。

```bash
git rev-list --count 4c7aab6..012de5f6
# 与 GitHub compare 的 ahead_by 相同（API: /repos/{owner}/{repo}/compare/{base}...{head}）
```

---

## 2. PowerShell 重定向写文件

**现象**：`Get-Content` 数出 293 行，`read` 工具报 368 行；据此判断"文件混了只有 CR 的换行"，实际是读取方式造成的假象。

**根因**：`>` 与 `Set-Content` 的默认编码随宿主变化（历史上写去过 UTF-16），控制台又按 GBK 解码 UTF-8 中文，于是换行、行数、字符全部错判。

**正确做法**：写文件用工具或 `node`；要判断编码/换行，直接读字节，不要经过 PowerShell 管道与重定向。

```bash
node -e "const b=require('fs').readFileSync('docs/README.md');console.log('BOM',b[0]===0xEF,'CRLF',(b.toString('latin1').match(/\r\n/g)||[]).length,'CR',(b.toString('latin1').match(/\r(?!\n)/g)||[]).length)"
```

**怎么验证**：以字节级统计为准，并与 `git diff` 对照——编码问题不该出现在 diff 里。

---

## 3. `.gitignore` 的 `dist/` 吞掉 `web/dist`

**现象**：本机 `go build` 通过，别人 clone 下来报 `web/web.go:9:12: pattern all:dist: no matching files found`。

**根因**：产物规则写成非锚定的 `dist/`，把内嵌前端目录 `web/dist` 一并忽略了。`//go:embed all:dist` 是编译期求值，所以错误在别人机器上才出现。

**正确做法**：产物规则锚定到仓库根；内嵌目录必须进库。

```gitignore
/dist/
/bin/
/build/
```

**怎么验证**：两条都要过，缺一不算。

```bash
git check-ignore -v web/dist/index.html   # 必须无输出、exit 1
# 按 git 真实可见文件集复制一棵树，在其中 go build
git ls-files -co --exclude-standard | ... ; go build ./...
```

---

## 4. 编译产物被 `git add .` 扫进库

**现象**：交叉编译出的 darwin 二进制留在工作区，随时可能被一把 `git add .` 带进仓库。

**根因**：产物与源码同目录，`git add .` 不区分来源。

**正确做法**：产物一律不入库，文档只给构建命令；`.gitignore` 覆盖常见产物形态。

```gitignore
*.exe
*.app
*.dmg
```

**怎么验证**：

```bash
git ls-files | grep -Ei '\.(exe|dmg)$|\.app/' ; # 必须无输出
git status --short --ignored | grep -iE 'bin/|build/'
```

---

## 5. `Page.Eval` 在主线程回调里调用

**现象**：在 `OnDidFinishLoad` 里调 `Eval`，进程卡住，约 25 秒后超时退出，日志只留一行 `did-finish-load`。

**根因**：`Eval` 自己需要主线程，而回调已经在主线程上，形成自锁。

**正确做法**：从 goroutine 里调用，再用 channel 或同步原语把结果送回测试协程。

**怎么验证**：同一断言写在 goroutine 里能稳定拿到值；连续跑多次不超时。

---

## 6. GUI 测试在没有桌面会话时永远红

**现象**：受限会话里 WebView2 控制器返回 `0x8000FFFF`，E2E 用例全红。

**根因**：WebView2 需要真实桌面会话；受限会话不具备。

**正确做法**：用 `MYGO_E2E=1` 作为开关，未设或无会话时 `t.Skip`，退出码 0。

**怎么验证**：两组结果都要有——设开关时 PASS，未设时 SKIP 且 `exit 0`；只有前者等于没验跳过逻辑。

```bash
MYGO_E2E=1 go test ./web -run 'TestM1|TestM2' -v
go test ./web -run 'TestM1|TestM2' -v    # 期望 SKIP + ok
```

---

## 7. ffmpeg 把绝对路径写进 m3u8

**现象**：播放器报多条 `fragLoadError code=0`，服务器一条分片请求都没收到。

**根因**：生成播放列表时用了绝对路径，`EXT-X-MAP` 里是 `C:\...\init.mp4`，hls.js 把它当成 URL 去请求，于是请求根本没发到服务端。

**正确做法**：在分片目录内用**相对路径**生成；排查时打印 hls.js 实际请求的地址，不要只看报错码。

```bash
cd <PROBE_DIR>/hls
ffmpeg ... -hls_fmp4_init_filename "init.mp4" -hls_segment_filename "seg%d.m4s" "index.m3u8"
```

**怎么验证**：监听 `Events.FRAG_LOADING` 打印 `frag.url`，路径应为相对或服务端地址，且服务端能看到对应请求。

---

## 8. 对着不存在的成本做设计

**现象**：因为"Safari 的 MSE 不吃 MPEG-TS"，准备在代理层加转封装，并把它写进了方案与成本。

**根因**：结论来自文档推断，没有实测。实测发现 hls.js 在 JS 内已把 TS 转封装成 fMP4 后才交给 MSE，WebView2 全程没见过 TS。

**正确做法**：成本判断先找一次最小实测，再决定要不要花钱；实测与文档冲突时以实测为准，并把它写进文档。

**怎么验证**：用同一段视频分别产出 fMP4 与 TS 两组分片，走同一代理、同一 hls.js，记录首帧与 seek；能在两种容器下正常播放即证伪。

---

## 9. 多人并行改前端时产物串味

**现象**：提交里带上了别人正在写的任务构建出的 `web/dist`，等于把一个任务的产物塞进另一个任务的提交。

**根因**：`web/dist` 是全仓共享的工作区状态，不受分支或文件级隔离保护。

**正确做法**：在临时 worktree 里只应用本条改动重建产物；**并在目标提交上复跑一次构建**，确认产物由该提交的源码产出。

```bash
git worktree add ../wt-<task> <target-commit>
# 只拷入本条改动的源码，构建
git status --porcelain web/dist    # 在目标提交上必须为空
```

**怎么验证**：`git status --porcelain web/dist` 为空，且重建 sha256 与提交内产物一致。

---

## 10. 同一个数值散在多个文件

**现象**：一个数改了六处漏五处，来回返工多轮，最后连"哪个是权威值"都说不清。

**根因**：把同一个事实复制到多份文档，每份都成了独立来源。

**正确做法**：数值只在唯一一处定义，其余位置只写引用或指向该处；能机器校验的就进脚本，别靠记性。

**怎么验证**：改那唯一一处，其余位置不需要改；用 grep/脚本统计是否还存在重复定义。

---

## 11. 文档里写绝对路径

**现象**：链接指向 `C:/Users/...`，在别人机器上点开是空的；审阅时也看出"文件里根本没有这个路径"。

**根因**：把本机的绝对路径当成通用引用贴进文档。

**正确做法**：一律用相对仓库根的路径；需要在对话里指路时，再用绝对路径。

**怎么验证**：`node scripts/check-docs.mjs` 第一项会解析所有相对链接，指向不存在的文件即失败。

---

## 维护规则

- 新坑与修复同一次提交登记，不许攒。
- 条目必须有可复现的验证方式；写不出验证的，先当成未定，不登记。
- 同一个坑第二次出现时，先补自动化（脚本或断言），再补文字。
