// api:1 宿主的本地实现：固定片单，不访问网络。M2 的页面只依赖 data/host 的接口。
import type {
  Host,
  HostCategories,
  HostDetail,
  HostItem,
  HostPage,
  HostSeason,
} from "./host";

const CATEGORIES: HostCategories = {
  items: [
    { id: "short", name: "短剧" },
    { id: "series", name: "剧集" },
    { id: "doc", name: "纪录片" },
  ],
};

const ITEMS: HostItem[] = [
  { id: "film-001", title: "深海列车", subtitle: "剧情 · 2024", episodeCount: 12 },
  { id: "film-002", title: "北纬三十度", subtitle: "纪录片 · 2023", episodeCount: 6 },
  { id: "film-003", title: "夜行线路", subtitle: "悬疑 · 2025", episodeCount: 12 },
  { id: "film-004", title: "夏末回信", subtitle: "短剧 · 2025", episodeCount: 24 },
];

// 每部片子的季与集；集 id 由片子 id 与集号拼成，便于详情页按集定位。
function seasonsOf(id: string, title: string, count: number): HostSeason[] {
  const episodes = Array.from({ length: count }, (_, index) => ({
    id: `${id}-ep-${index + 1}`,
    title: `第 ${index + 1} 集`,
    durationMs: 6 * 60 * 1000,
  }));
  return [
    { id: `${id}-s1`, title: `${title} 第 1 季`, episodes },
    { id: `${id}-s2`, title: `${title} 第 2 季`, episodes: episodes.slice(0, Math.max(1, count - 6)) },
  ];
}

const DETAILS = new Map<string, HostDetail>(
  ITEMS.map((item) => [
    item.id,
    {
      id: item.id,
      title: item.title,
      description: `${item.title}：${item.subtitle ?? ""}。这里放简介，详情页按行数折叠后展开全文。`,
      seasons: seasonsOf(item.id, item.title, item.episodeCount ?? 4),
    },
  ]),
);

// 条目归属的分类；分类 id 与 categories() 的返回对应。
const CATEGORY_OF: Record<string, string> = {
  "film-001": "series",
  "film-002": "doc",
  "film-003": "series",
  "film-004": "short",
};

// 遍历顺序固定，页面断言不依赖随机顺序。
function pageOf(items: HostItem[], page: number, size: number): HostPage {
  const start = Math.max(0, (page - 1) * size);
  const slice = items.slice(start, start + size);
  return { items: slice, hasMore: start + size < items.length };
}

// fixture 的故障注入：测试置位 window.__forceHostError 后，取目录会抛错，
// 用来验证页面的错误态与重试，不涉及网络。
function guard(): void {
  if ((window as unknown as { __forceHostError?: boolean }).__forceHostError) {
    throw new Error("fixture failure");
  }
}

// 宿主调用次数，供页面测试核对缓存是否真正命中。
function bump(kind: "catalog" | "search"): void {
  const host = window as unknown as { __hostCalls?: { catalog: number; search: number } };
  host.__hostCalls ??= { catalog: 0, search: 0 };
  host.__hostCalls[kind] += 1;
}

const MODULE = { id: "hongguo", version: "1.0.4" };

export const host: Host = {
  categories: () => CATEGORIES,
  catalogPage: (page, category) => {
    guard();
    bump("catalog");
    const filtered =
      category === "" ? ITEMS : ITEMS.filter((item) => CATEGORY_OF[item.id] === category);
    return { ...pageOf(filtered, page, 12), warning: filtered.length === 0 ? "该分类暂无内容" : undefined };
  },
  search: (query) => {
    bump("search");
    const keyword = query.trim();
    if (keyword === "") {
      return { items: [], hasMore: false };
    }
    const hits = ITEMS.filter((item) => item.title.includes(keyword));
    return { ...pageOf(hits, 1, 12), warning: hits.length === 0 ? "没有匹配的结果" : undefined };
  },
  moduleVersion: () => MODULE.version,
  detail: (sourceId) => DETAILS.get(sourceId) ?? null,
  // 取流属于播放器，M2 不消费；留接口是为了与 api:1 的形状一致。
  resolveMedia: () => {
    throw new Error("M2 只做内容页，取流由播放器接入");
  },
};
