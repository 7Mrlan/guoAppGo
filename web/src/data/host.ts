// 站源模块 api:1 的宿主接口形状。
//
// 形状取自上游 DEVELOPMENT.md 的「站源模块」一节：main.js 用 CommonJS 导出
// categories(force)、catalogPage(page, category)、detail(sourceId)、
// resolveMedia(task)，可选 validCategory(category)；catalogPage 与 search 返回
// { items, hasMore, warning }；这些函数一律同步返回。

export interface HostCategory {
  id: string;
  name: string;
}

export interface HostItem {
  id: string;
  title: string;
  cover?: string;
  subtitle?: string;
  episodeCount?: number;
}

export interface HostPage {
  items: HostItem[];
  hasMore: boolean;
  warning?: string;
}

export interface HostEpisode {
  id: string;
  title: string;
  durationMs?: number;
}

export interface HostSeason {
  id: string;
  title: string;
  episodes: HostEpisode[];
}

export interface HostDetail {
  id: string;
  title: string;
  description: string;
  seasons: HostSeason[];
}

export interface HostCategories {
  items: HostCategory[];
  incomplete?: boolean;
}

export interface Host {
  categories(force?: boolean): HostCategories;
  catalogPage(page: number, category: string): HostPage;
  search(query: string): HostPage;
  detail(sourceId: string): HostDetail | null;
  resolveMedia(task: unknown): never;
}
