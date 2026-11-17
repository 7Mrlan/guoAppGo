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
  const requiredTags = ["app-v2.0.1-113", "v0.9.2", "v0.1.1"];

  if (!fs.existsSync(versionsFile)) {
    problems.push("docs/versions.md 不存在");
  } else {
    const text = fs.readFileSync(versionsFile, "utf8");
    for (const m of text.matchAll(/\b[0-9a-f]{40}\b/g)) shaSet.add(m[0]);
    if (shaSet.size < 3) problems.push(`docs/versions.md: 只找到 ${shaSet.size} 个 40 位 sha，少于 3 条基准`);
    for (const tag of requiredTags) {
      if (!text.includes(tag)) problems.push(`docs/versions.md: 缺少版本 tag ${tag}`);
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

/** 串行执行四项检查并打印结果；有失败项时以退出码 1 结束，供提交前门禁调用。 */
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
