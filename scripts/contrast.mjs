// Prints the WCAG contrast of the foreground/background pairs the design
// tokens promise, for the dark and the light theme, and fails when a pair
// drops below its threshold. Read-only: it never writes the built frontend.
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import { blockTokens } from "./check-tokens.mjs";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const tokensPath = join(root, "web", "src", "tokens.css");

// 12 pairs: body text must reach 4.5, large text and UI accents 3.0.
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

// 读取 tokens.css，分别检查暗色与亮色，全部达标才退出 0。
function main() {
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
  if (failures.length > 0) {
    console.log(`FAIL  对比度  ${failures.length} 组不达标`);
    for (const failure of failures) {
      console.log(`        - ${failure}`);
    }
    process.exit(1);
  }
  console.log(`PASS  对比度  ${PAIRS.length} 组 × 2 主题，全部达标`);
}

if (process.argv[1] && process.argv[1].endsWith("contrast.mjs")) {
  main();
}
