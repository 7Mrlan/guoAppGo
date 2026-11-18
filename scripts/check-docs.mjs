#!/usr/bin/env node
/**
 * 文档一致性检查：链接与文件形态、章节编号连续性、版本基准确性与一致、README 登记。
 * 用法：node scripts/check-docs.mjs；退出码 0 全过，1 有失败项。无外部依赖。
 */

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const ROOT = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const DOCS = path.join(ROOT, "docs");

/** 汇总每项检查的结论与明细，供 main 统一输出。 */
const results = [];
/** 失败项计数，决定进程退出码；非零即视为文档不合格。 */
let failed = 0;

// 章节编号要求：每份文件从 start 连续编到 end，且存在指定附录标题
const numberingRules = [
  { file: "docs/evaluation.md", start: 0, end: 4, appendix: "附录 A" },
  { file: "docs/evaluation-plan.md", start: 5, end: 8, appendix: "附录 B" },
];

// README 不必逐个登记的文件
const registrationWhitelist = ["docs/README.md", "docs/adr/0000-template.md"];

/** 记录一项检查结果并累计失败数，供 main 决定退出码。 */
function record(name, ok, detail) {
  results.push({ name, ok, detail });
  if (!ok) failed++;
}

/** 递归收集目录下的所有 .md 文件；目录不存在时返回空数组。 */
function walk(dir, out = []) {
  if (!fs.existsSync(dir)) return out;
  for (const e of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, e.name);
    if (e.isDirectory()) walk(p, out);
    else if (e.name.endsWith(".md")) out.push(p);
  }
  return out;
}

/** 把绝对路径转成相对仓库根的 / 分隔路径，用于输出与 README 登记比对。 */
function toRepoPath(p) {
  return path.relative(ROOT, p).split(path.sep).join("/");
}

/** 检查一：所有 md 的相对链接可解析，且文件为 UTF-8、无 BOM、纯 LF、末尾有换行。 */
function checkLinksAndForm(allFiles) {
  const problems = [];
  let linkCount = 0;
  for (const file of allFiles) {
    const raw = fs.readFileSync(file);
    let text;
    try {
      text = new TextDecoder("utf-8", { fatal: true }).decode(raw);
    } catch {
      problems.push(`${toRepoPath(file)}: 不是合法 UTF-8`);
      continue;
    }
    if (raw[0] === 0xef && raw[1] === 0xbb) problems.push(`${toRepoPath(file)}: 含 UTF-8 BOM`);
    const latin = raw.toString("latin1");
    if (/\r\n/.test(latin)) problems.push(`${toRepoPath(file)}: 含 CRLF，要求纯 LF`);
    if (/\r(?!\n)/.test(latin)) problems.push(`${toRepoPath(file)}: 含 CR-only 换行`);
    if (raw.length && raw[raw.length - 1] !== 0x0a) problems.push(`${toRepoPath(file)}: 文件末尾缺换行`);

    const lines = text.split("\n");
    // 只校验相对链接的目标存在性；外链、锚点不跟
    lines.forEach((line, i) => {
      const re = /\[[^\]]*\]\(([^)\s]+)\)/g;
      let m;
      while ((m = re.exec(line))) {
        const target = m[1];
        if (/^(https?:|mailto:|#)/.test(target)) continue;
        linkCount++;
        const clean = decodeURI(target.split("#")[0]);
        if (!clean) continue;
        const resolved = path.resolve(path.dirname(file), clean);
        if (!fs.existsSync(resolved)) {
          problems.push(`${toRepoPath(file)}:${i + 1}: 链接目标不存在 -> ${target}`);
        }
      }
    });
  }
  const detail = problems.length ? problems : `检查 ${linkCount} 条内链，${allFiles.length} 个 md`;
  record("链接与文件形态", problems.length === 0, detail);
}

/** 检查二：评估两份的二级标题编号必须连续，且含规定附录标题，防止拆分后编号重起或跳号。 */
function checkHeadingNumbering() {
  const problems = [];
  for (const rule of numberingRules) {
    const abs = path.join(ROOT, rule.file);
    if (!fs.existsSync(abs)) {
      problems.push(`${rule.file}: 文件不存在`);
      continue;
    }
    const text = fs.readFileSync(abs, "utf8");
    const got = [];
    for (const line of text.split("\n")) {
      const m = line.match(/^##\s+(\d+)\.\s+/);
      if (m) got.push(Number(m[1]));
    }
    const want = [];
    for (let n = rule.start; n <= rule.end; n++) want.push(n);
    const same = got.length === want.length && got.every((v, i) => v === want[i]);
    if (!same) problems.push(`${rule.file}: 章节编号应为 ${want.join(",")}，实际 ${got.join(",") || "无"}`);
    if (!text.includes(`## ${rule.appendix}`)) problems.push(`${rule.file}: 缺少标题 "${rule.appendix}"`);
  }
  const detail = problems.length
    ? problems
    : numberingRules.map((r) => `${r.file} ${r.start}-${r.end}`).join("；");
  record("章节编号连续性", problems.length === 0, detail);
}

/** 检查三：versions.md 是唯一事实来源；引用 40 位 sha 的文档必须有基准节、回链 versions.md，且 sha 不出界。 */
function checkVersionBaseline() {
  const problems = [];
  const versionsFile = path.join(DOCS, "versions.md");
  const shaSet = new Set();
  const pinnedTags = ["v0.9.2", "v0.1.1"];

  if (!fs.existsSync(versionsFile)) {
    problems.push("docs/versions.md 不存在");
  } else {
    const text = fs.readFileSync(versionsFile, "utf8");
    for (const m of text.matchAll(/\b[0-9a-f]{40}\b/g)) shaSet.add(m[0]);
    if (shaSet.size < 3) problems.push(`docs/versions.md: 只找到 ${shaSet.size} 个 40 位 sha，少于 3 条基准`);
    // mygo 与 moejs 钉死在 tag；原应用跟 main 提交，不跟 tag
    for (const tag of pinnedTags) {
      if (!text.includes(tag)) problems.push(`docs/versions.md: 缺少钉死版本 tag ${tag}`);
    }
    // 原应用那一行必须是 main 提交号，写 tag 即视为口径过期
    const guoRow = text.split("\n").find((l) => l.includes("Xanerix/guoapp") && l.trim().startsWith("|"));
    if (!guoRow) {
      problems.push("docs/versions.md: 未找到 Xanerix/guoapp 的基准行");
    } else {
      if (!/\b[0-9a-f]{40}\b/.test(guoRow)) problems.push("docs/versions.md: guoapp 基准行缺少 40 位提交号");
      if (!/main|提交/.test(guoRow)) problems.push("docs/versions.md: guoapp 基准行需标明以 main 提交为准");
    }
    if (!text.includes("app-v2.0.1-113")) {
      problems.push("docs/versions.md: 缺少 app-v2.0.1-113 阅读锚点（evaluation.md 的实测数据来源）");
    }
  }

  for (const file of ["docs/evaluation.md", "docs/spike-stack.md"]) {
    const abs = path.join(ROOT, file);
    if (!fs.existsSync(abs)) {
      problems.push(`${file}: 文件不存在`);
      continue;
    }
    const text = fs.readFileSync(abs, "utf8");
    if (!/^##\s+0\.\s+版本基准/m.test(text)) problems.push(`${file}: 缺少"## 0. 版本基准"节`);
    for (const m of text.matchAll(/\b[0-9a-f]{40}\b/g)) {
      if (shaSet.size && !shaSet.has(m[0])) problems.push(`${file}: 出现 versions.md 之外的 sha ${m[0].slice(0, 12)}…`);
    }
    if (!/\]\(\.?\.?\/?versions\.md\)/.test(text)) {
      problems.push(`${file}: 未链接 versions.md（唯一事实来源）`);
    }
  }

  // 代码侧版本必须与基准表同源：go.mod 的 mygo 与 package-lock 的 mygo-cli 要一致且都写进 versions.md
  const versionsText = fs.existsSync(versionsFile) ? fs.readFileSync(versionsFile, "utf8") : "";
  const goModPath = path.join(ROOT, "go.mod");
  // go.mod 中 mygo 的模块版本，形如 v0.9.2
  const goMygoMatch = fs.existsSync(goModPath)
    ? fs.readFileSync(goModPath, "utf8").match(/github\.com\/egoist\/mygo\s+(v\d+\.\d+\.\d+\S*)/)
    : null;
  if (!goMygoMatch) problems.push("go.mod: 未找到 github.com/egoist/mygo 的模块版本");
  if (goMygoMatch && !versionsText.includes(goMygoMatch[1])) {
    problems.push(`go.mod 的 mygo ${goMygoMatch[1]} 未写进 versions.md`);
  }
  if (!versionsText.includes("mygo-cli")) problems.push("versions.md: 未登记 npm 侧 mygo-cli");
  const lockPath = path.join(ROOT, "package-lock.json");
  if (fs.existsSync(lockPath)) {
    // package-lock 里 mygo-cli 的实装版本，不带 v 前缀
    const lockVersion = (JSON.parse(fs.readFileSync(lockPath, "utf8")).packages || {})["node_modules/mygo-cli"]?.version;
    if (!lockVersion) problems.push("package-lock.json: 未找到 node_modules/mygo-cli");
    else if (goMygoMatch && lockVersion !== goMygoMatch[1].replace(/^v/, "")) {
      problems.push(`mygo 版本分叉：go.mod ${goMygoMatch[1]} / package-lock mygo-cli ${lockVersion}`);
    }
  }
  const detail = problems.length ? problems : `基准 ${shaSet.size} 条 sha，引用一致`;
  record("版本基准确性与一致", problems.length === 0, detail);
}

/** 检查四：README 登记与 docs/ 实际文件双向一致，新增文件不登记即失败。 */
function checkReadmeRegistration(docsFiles) {
  const problems = [];
  const readme = path.join(DOCS, "README.md");
  if (!fs.existsSync(readme)) {
    record("README 登记", false, ["docs/README.md 不存在"]);
    return;
  }

  const text = fs.readFileSync(readme, "utf8");
  const mentioned = new Set();
  // 目录树代码块：每行 "相对路径.md  说明"，基准是 docs/
  let inFence = false;
  for (const line of text.split("\n")) {
    if (/^\s*```/.test(line)) {
      inFence = !inFence;
      continue;
    }
    if (!inFence) continue;
    const m = line.trim().match(/^([A-Za-z0-9_./-]+\.md)\s/);
    if (m) mentioned.add("docs/" + m[1].replace(/^docs\//, ""));
  }
  // 反引号里的 docs/... 与相对链接两种写法都算登记
  for (const m of text.matchAll(/`(docs\/[A-Za-z0-9_./-]+\.md)`/g)) mentioned.add(m[1]);
  for (const m of text.matchAll(/\]\(([^)\s]+\.md)\)/g)) {
    const resolved = path.resolve(DOCS, decodeURI(m[1].split("#")[0]));
    if (resolved.startsWith(DOCS)) mentioned.add(toRepoPath(resolved));
  }

  for (const f of mentioned) {
    if (!fs.existsSync(path.join(ROOT, f))) problems.push(`docs/README.md 登记了不存在的文件: ${f}`);
  }
  for (const abs of docsFiles) {
    const r = toRepoPath(abs);
    if (registrationWhitelist.includes(r)) continue;
    if (!mentioned.has(r)) problems.push(`docs/ 下有文件未在 README 登记: ${r}`);
  }
  const detail = problems.length ? problems : `登记 ${mentioned.size} 个文件，与 docs/ 实际一致`;
  record("README 登记", problems.length === 0, detail);
}

/** 检查五：人天口径只在 evaluation-plan.md §6.3 定义；其他文件的区间与"对外报"必须与它一致，否则视为漂移。 */
function checkBudgetNumbers(allFiles) {
  const problems = [];
  const planPath = path.join(ROOT, "docs/evaluation-plan.md");
  if (!fs.existsSync(planPath)) {
    record("人天口径一致", false, ["docs/evaluation-plan.md 不存在，无法取口径"]);
    return;
  }
  const planText = fs.readFileSync(planPath, "utf8");
  // 口径行：v1 实际（M6-lite）那一行，同时含小计、合计与对外值
  const canonLine = planText.split("\n").find((l) => l.includes("M6-lite") && l.includes("小计") && l.includes("合计"));
  // 口径区间：小计与合计各自的下界/上界
  const sub = canonLine && canonLine.match(/小计\s*\**\s*([\d.]+)~([\d.]+)/);
  const tot = canonLine && canonLine.match(/合计\s*\**\s*([\d.]+)~([\d.]+)/);
  // 对外报的保守值，必须是单一数字
  const out = canonLine && canonLine.match(/对外[^0-9]{0,8}([0-9]+)/);
  if (!sub || !tot || !out) {
    problems.push("§6.3 口径行解析失败：需同时含“小计 X~Y”“合计 X~Y”“对外… N”");
    record("人天口径一致", false, problems);
    return;
  }
  const allowed = new Set([`${sub[1]}~${sub[2]}`, `${tot[1]}~${tot[2]}`]);
  // 失败路径与多源存档口径也允许出现，只取路线级（下界 ≥50）的区间，避免把模块级估算放进来
  for (const line of planText.split("\n")) {
    if (!/失败|退路|多源/.test(line)) continue;
    for (const m of line.matchAll(/([\d.]+)~([\d.]+)\s*人天/g)) {
      if (Number(m[1]) >= 50) allowed.add(`${m[1]}~${m[2]}`);
    }
  }
  const outValue = out[1];
  // 允许出现的其他口径，但所在行必须带限定词，说明它不是 v1 主线
  const qualifiers = ["备选", "退路", "失败", "多源", "旧基数", "仅存档", "若 R1", "不通过", "R1 不", "更省", "差额", "多 ", "少 "];

  for (const file of allFiles) {
    const text = fs.readFileSync(file, "utf8");
    const lines = text.split("\n");
    lines.forEach((line, i) => {
      const at = `${toRepoPath(file)}:${i + 1}`;
      // 只校验路线级（B 单源 / v1）行上的区间；模块级估算（如 20~30、5~8）不在此列
      const routeLine = /(^|[^A-Za-z])B([^A-Za-z]|$)/.test(line) && /单源|v1|对外/.test(line);
      if (routeLine) {
        for (const m of line.matchAll(/([\d.]+)~([\d.]+)\s*人天/g)) {
          // 两端都不到 100 的视为增量、加价或差值（如 20~30、5~8、3~17），不参与总口径比对
          if (Number(m[1]) < 100 || Number(m[2]) < 100) continue;
          const range = `${m[1]}~${m[2]}`;
          if (allowed.has(range)) continue;
          if (qualifiers.some((q) => line.includes(q))) continue;
          problems.push(`${at}: 人天区间 ${range} 与 §6.3 口径不符（允许 ${[...allowed].join(" / ")}）`);
        }
      }
      // 对外口径：只允许一个数，全库校验
      for (const m of line.matchAll(/对外[^0-9]{0,8}([0-9]+)\s*人天/g)) {
        if (m[1] !== outValue) problems.push(`${at}: 对外报 ${m[1]} 人天，与 §6.3 的 ${outValue} 不符`);
      }
    });
  }
  const detail = problems.length ? problems : `口径 ${[...allowed].join(" / ")}，对外 ${outValue}`;
  record("人天口径一致", problems.length === 0, detail);
}

/** 串行执行五项检查并打印结果；有失败项时以退出码 1 结束，供提交前门禁调用。 */
function main() {
  const docsFiles = walk(DOCS);
  const rootFiles = ["README.md", "CHANGELOG.md"]
    .map((f) => path.join(ROOT, f))
    .filter((p) => fs.existsSync(p));
  const allFiles = [...rootFiles, ...docsFiles];

  checkLinksAndForm(allFiles);
  checkHeadingNumbering();
  checkVersionBaseline();
  checkReadmeRegistration(docsFiles);
  checkBudgetNumbers(allFiles);

  console.log(`文档检查（${results.length} 项）：`);
  for (const r of results) {
    const head = `${r.ok ? "PASS" : "FAIL"}  ${r.name}`;
    if (Array.isArray(r.detail)) {
      console.log(head);
      for (const d of r.detail) console.log(`        - ${d}`);
    } else {
      console.log(`${head}  （${r.detail}）`);
    }
  }
  if (failed) {
    console.log(`\n${failed} 项失败。`);
    process.exit(1);
  }
  console.log("\n全部通过。");
}

main();
