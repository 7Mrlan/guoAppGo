// Prints the WCAG contrast of the foreground/background pairs the design
// tokens promise, for the dark and the light theme, and fails when a pair
// drops below its threshold or below the stored baseline. Never writes the
// built frontend; --update is the only writer, and only for the baseline file.
import { readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { blockTokens } from "./check-tokens.mjs";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const tokensPath = join(root, "web", "src", "tokens.css");
const baselinePath = join(root, "scripts", "contrast-baseline.json");

// 14 pairs: body text must reach 4.5, large text, UI accents and control
// borders 3.0.
const PAIRS = [
  ["--color-text", "--color-bg", 4.5],
  ["--color-text", "--color-surface", 4.5],
  ["--color-text-muted", "--color-bg", 4.5],
  ["--color-text-muted", "--color-surface", 4.5],
  ["--color-accent", "--color-bg", 4.5],
  ["--color-text", "--color-surface-raised", 4.5],
  ["--color-accent-text", "--color-accent", 3.0],
  ["--color-danger", "--color-bg", 3.0],
  ["--color-success", "--color-bg", 3.0],
  ["--color-accent", "--color-surface-raised", 3.0],
  ["--color-danger", "--color-surface", 3.0],
  ["--color-success", "--color-surface", 3.0],
  ["--color-border-strong", "--color-bg", 3.0],
  ["--color-border-strong", "--color-surface", 3.0],
];

/** Parses #rgb, #rrggbb and #rrggbbaa into linear sRGB channels. */
export function parseColor(value) {
  const hex = value.trim().replace(/^#/, "");
  if (!/^[0-9a-f]{3,8}$/i.test(hex)) {
    return null;
  }
  const expand = (s) => (s.length === 1 ? s + s : s);
  const parts =
    hex.length === 3 || hex.length === 4
      ? [...hex].map(expand)
      : [hex.slice(0, 2), hex.slice(2, 4), hex.slice(4, 6)];
  const channel = (s) => {
    const srgb = parseInt(s, 16) / 255;
    return srgb <= 0.04045 ? srgb / 12.92 : ((srgb + 0.055) / 1.055) ** 2.4;
  };
  return parts.slice(0, 3).map(channel);
}

/** WCAG relative luminance of linear sRGB channels. */
export function luminance([r, g, b]) {
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** WCAG contrast ratio between two colours. */
export function contrast(fg, bg) {
  const a = luminance(fg);
  const b = luminance(bg);
  const [light, dark] = a > b ? [a, b] : [b, a];
  return (light + 0.05) / (dark + 0.05);
}

// 取出一个选择器块里的令牌；块不存在时抛错。
function themeTokens(css, selector) {
  const tokens = blockTokens(css, selector);
  if (tokens === null) {
    throw new Error(`缺少 ${selector} 块`);
  }
  return tokens;
}

/** Stable identity of one pair in one theme, as used in the baseline file. */
function pairKey(name, fg, bg) {
  return `${name}|${fg}|${bg}`;
}

/** Reads the stored ratios, or null when the baseline was never generated. */
function readBaseline() {
  try {
    return JSON.parse(readFileSync(baselinePath, "utf8"));
  } catch (err) {
    if (err.code === "ENOENT") {
      return null;
    }
    throw err;
  }
}

/** Writes the ratios sorted by key, so the file changes only when a ratio does. */
function writeBaseline(rows) {
  const stored = {};
  const sorted = [...rows].sort((a, b) =>
    pairKey(a.name, a.fg, a.bg) < pairKey(b.name, b.fg, b.bg) ? -1 : 1,
  );
  for (const row of sorted) {
    stored[pairKey(row.name, row.fg, row.bg)] = Number(row.ratio.toFixed(2));
  }
  writeFileSync(baselinePath, `${JSON.stringify(stored, null, 2)}\n`);
}

/**
 * Compares the measured ratios with the baseline: an existing pair may not
 * drop, a pair missing from the baseline must be stored first, and a stored
 * pair that no longer exists in PAIRS is stale. Ratios within 0.01 count as
 * unchanged, since they are stored at two decimals.
 */
function compareBaseline(rows, baseline) {
  const failures = [];
  const keys = new Set();
  for (const row of rows) {
    const key = pairKey(row.name, row.fg, row.bg);
    keys.add(key);
    const stored = baseline[key];
    if (stored === undefined) {
      failures.push(`新配对未入基线：${key}，先跑 node scripts/contrast.mjs --update`);
      continue;
    }
    if (row.ratio + 0.01 < stored) {
      failures.push(`低于基线：${key} 现值 ${row.ratio.toFixed(2)} vs 基线 ${stored.toFixed(2)}`);
    }
  }
  for (const key of Object.keys(baseline)) {
    if (!keys.has(key)) {
      failures.push(`基线残留：${key}`);
    }
  }
  return failures;
}

// 按 PAIRS 逐对计算某主题的对比度，返回明细行与不达标项。
function checkTheme(name, base, overrides) {
  const failures = [];
  const rows = [];
  for (const [fgName, bgName, min] of PAIRS) {
    const fgValue = overrides.get(fgName) ?? base.get(fgName);
    const bgValue = overrides.get(bgName) ?? base.get(bgName);
    const fg = parseColor(fgValue ?? "");
    const bg = parseColor(bgValue ?? "");
    if (!fg || !bg) {
      failures.push(`${name} ${fgName} on ${bgName}: 无法解析颜色`);
      continue;
    }
    const ratio = contrast(fg, bg);
    const ok = ratio >= min;
    if (!ok) {
      failures.push(`${name} ${fgName} on ${bgName}: ${ratio.toFixed(2)} < ${min}`);
    }
    rows.push({ name, fg: fgName, bg: bgName, ratio, min, ok });
  }
  return { rows, failures };
}

// 读取 tokens.css，分别检查暗色与亮色，再与已存基线比对。只有 --update 会写基线。
function main(argv) {
  const update = argv.includes("--update");

  const css = readFileSync(tokensPath, "utf8");
  const base = themeTokens(css, ":root");
  const light = new Map(themeTokens(css, '[data-theme="light"]'));

  const { rows, failures } = checkTheme("dark", base, new Map());
  const lightResult = checkTheme("light", base, light);
  failures.push(...lightResult.failures);
  rows.push(...lightResult.rows);

  for (const row of rows) {
    const ratio = row.ratio.toFixed(2).padStart(5);
    const verdict = row.ok ? "ok" : "FAIL";
    console.log(`${row.name.padEnd(5)} ${row.fg.padEnd(22)} on ${row.bg.padEnd(22)} ${ratio}  min ${row.min}  ${verdict}`);
  }

  if (update) {
    if (failures.length > 0) {
      console.log("FAIL  基线未写入：有配对低于自身阈值，先把颜色调到达标");
      for (const failure of failures) {
        console.log(`        - ${failure}`);
      }
      process.exit(1);
    }
    writeBaseline(rows);
    console.log(`UPDATE  对比度基线  scripts/contrast-baseline.json  ${rows.length} 条`);
    return;
  }

  const baseline = readBaseline();
  if (baseline === null) {
    failures.push("缺少基线文件 scripts/contrast-baseline.json，先跑 node scripts/contrast.mjs --update");
  } else {
    failures.push(...compareBaseline(rows, baseline));
  }

  if (failures.length > 0) {
    console.log(`FAIL  对比度  ${failures.length} 项不合格`);
    for (const failure of failures) {
      console.log(`        - ${failure}`);
    }
    process.exit(1);
  }
  console.log(`PASS  对比度  ${PAIRS.length} 组 × 2 主题，全部达标且不低于基线`);
}

if (process.argv[1] && process.argv[1].endsWith("contrast.mjs")) {
  main(process.argv.slice(2));
}
