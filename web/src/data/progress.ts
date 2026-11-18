// 观看进度：本地记录，M2 只读展示，写入留给播放器接入时补。
export interface ProgressEntry {
  id: string;
  episodeId: string;
  positionMs: number;
  updatedAt: string;
}

const ENTRIES: ProgressEntry[] = [
  { id: "film-001", episodeId: "film-001-ep-3", positionMs: 96_000, updatedAt: "2026-11-17T10:00:00Z" },
];

// 按更新时间倒序返回，页面直接用于"继续观看"。
export function recentProgress(): ProgressEntry[] {
  return [...ENTRIES].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt));
}
