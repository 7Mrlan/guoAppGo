import { build } from "vite";
import { createHash } from "node:crypto";
import { mkdir, readFile, readdir, rename, rm } from "node:fs/promises";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

// Builds the frontend into web/dist, which go:embed packs into the binary.
//
// Vite writes to web/.dist-tmp first; only a successful build replaces
// web/dist, so a broken toolchain or a failed npm ci can never ship an empty
// page.
const root = dirname(dirname(fileURLToPath(import.meta.url)));
const outDir = join(root, "web", "dist");
const tmpDir = join(root, "web", ".dist-tmp");
const oldDir = join(root, "web", ".dist-old");

/** Lists files under dir, relative to it, sorted, so the hash is stable. */
async function listFiles(dir, prefix = "") {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = [];
  for (const entry of entries.sort((a, b) => a.name.localeCompare(b.name))) {
    const rel = prefix ? `${prefix}/${entry.name}` : entry.name;
    if (entry.isDirectory()) {
      files.push(...(await listFiles(join(dir, entry.name), rel)));
    } else {
      files.push(rel);
    }
  }
  return files;
}

/** Hashes the directory's contents so two builds can be compared byte for byte. */
async function hashDir(dir) {
  const hash = createHash("sha256");
  for (const rel of await listFiles(dir)) {
    hash.update(rel);
    hash.update("\0");
    hash.update(await readFile(join(dir, rel)));
    hash.update("\0");
  }
  return hash.digest("hex");
}

/** Replaces web/dist with tmpDir; a failure at any step restores the previous output. */
async function swap() {
  await rm(oldDir, { recursive: true, force: true });
  let movedOld = false;
  try {
    await rename(outDir, oldDir);
    movedOld = true;
  } catch (err) {
    if (err.code !== "ENOENT") throw err;
  }
  try {
    await rename(tmpDir, outDir);
  } catch (err) {
    if (movedOld) await rename(oldDir, outDir);
    throw err;
  }
  await rm(oldDir, { recursive: true, force: true });
}

/** Builds the frontend, then swaps the result in and prints its hash. */
async function main() {
  await rm(tmpDir, { recursive: true, force: true });
  try {
    await build({ configFile: join(root, "vite.config.ts") });
    await mkdir(tmpDir, { recursive: true });
    await swap();
  } finally {
    await rm(tmpDir, { recursive: true, force: true });
  }
  console.log(`built web/dist sha256=${await hashDir(outDir)}`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
