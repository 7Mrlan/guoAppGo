// 目录缓存：键含站源模块版本，模块升级后旧缓存自然失效。
import { host } from "./fixture";
import type { HostPage } from "./host";

const cache = new Map<string, HostPage>();

// 模块版本由 fixtures 提供，测试可改写以模拟升级。
function moduleVersion(): string {
  const override = (window as unknown as { __fixtureModuleVersion?: string }).__fixtureModuleVersion;
  return override ?? host.moduleVersion();
}

export function cacheKey(category: string, page: number): string {
  return `${moduleVersion()}|${category}|${page}`;
}

// 命中缓存不发请求；未命中才调用宿主并记入缓存。
export function readCatalog(category: string, page: number): HostPage {
  const key = cacheKey(category, page);
  const hit = cache.get(key);
  if (hit !== undefined) {
    return hit;
  }
  const fresh = host.catalogPage(page, category);
  cache.set(key, fresh);
  return fresh;
}
