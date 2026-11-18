// Counts the design tokens of web/src/tokens.css and fails when the count or
// the coverage of the light theme drifts. Read-only: it never writes the
// built frontend, so it stays out of npm run build.
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const tokensPath = join(root, "web", "src", "tokens.css");

/** The 25 design tokens tokens.css must declare; changing this count changes the spec. */
export const EXPECTED_TOKENS = 25;

/** Returns the declarations of one selector block, mapped to trimmed values. */
export function blockTokens(css, selector) {
  const start = css.indexOf(`${selector} {`);
  if (start === -1) {
    return null;
  }
  const end = css.indexOf("}", start);
  if (end === -1) {
    return null;
  }
  const body = css.slice(start + selector.length + 2, end);
  const tokens = new Map();
  for (const line of body.split("\n")) {
    const match = line.match(/^\s*(--[a-z0-9-]+)\s*:\s*([^;]+);/i);
    if (match) {
      tokens.set(match[1], match[2].trim());
    }
  }
  return tokens;
}

// 校验 :root 的令牌数与亮色主题的覆盖完整性，返回问题列表。
export function checkTokens(css) {
  const problems = [];
  const base = blockTokens(css, ":root");
  if (base === null) {
    return [":root 块不存在"];
  }
  if (base.size !== EXPECTED_TOKENS) {
    problems.push(`令牌数 ${base.size}，期望 ${EXPECTED_TOKENS}`);
  }

  const light = blockTokens(css, '[data-theme="light"]');
  if (light === null) {
    problems.push('[data-theme="light"] 块不存在');
  } else {
    const missing = [...base.keys()].filter((name) => name.startsWith("--color-") && !light.has(name));
    if (missing.length > 0) {
      problems.push(`亮色主题缺少：${missing.join(", ")}`);
    }
    const unknown = [...light.keys()].filter((name) => !name.startsWith("--color-"));
    if (unknown.length > 0) {
      problems.push(`亮色主题只应覆盖颜色令牌：${unknown.join(", ")}`);
    }
  }
  return problems;
}

// 读取 tokens.css 并打印结果；有问题时以非零退出码结束。
function main() {
  const css = readFileSync(tokensPath, "utf8");
  const problems = checkTokens(css);
  const count = blockTokens(css, ":root").size;
  if (problems.length > 0) {
    console.log("FAIL  设计令牌");
    for (const problem of problems) {
      console.log(`        - ${problem}`);
    }
    process.exit(1);
  }
  console.log(`PASS  设计令牌  ${count} 个，亮色覆盖完整`);
}

if (process.argv[1] && process.argv[1].endsWith("check-tokens.mjs")) {
  main();
}
