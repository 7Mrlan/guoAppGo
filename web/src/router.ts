// hash 路由：路径到视图的映射，含带参数的路由。
export type RouteName =
  | "home"
  | "explore"
  | "search"
  | "settings"
  | "components"
  | "follow-list"
  | "mine"
  | "detail"
  | "follow"
  | "not-found";

export interface ParsedRoute {
  name: RouteName;
  // 带参数路由的取值，例如 detail 的 id。
  params: Record<string, string>;
}

const STATIC: Record<string, RouteName> = {
  "": "home",
  "/": "home",
  "/explore": "explore",
  "/search": "search",
  "/settings": "settings",
  "/dev/components": "components",
  "/follow": "follow-list",
  "/mine": "mine",
};

// 带参数路由：前缀与参数名的对应关系。
const PARAM_ROUTES: Array<{ prefix: string; name: RouteName; param: string }> = [
  { prefix: "/detail/", name: "detail", param: "id" },
  { prefix: "/follow/", name: "follow", param: "id" },
];

// 解析 hash；未知路径落到 not-found，参数按 URI 解码。
export function parseRoute(hash: string = location.hash): ParsedRoute {
  const path = hash.replace(/^#/, "").split("?")[0];
  const stat = STATIC[path];
  if (stat !== undefined) {
    return { name: stat, params: {} };
  }
  for (const route of PARAM_ROUTES) {
    const rest = path.slice(route.prefix.length);
    if (path.startsWith(route.prefix) && rest !== "" && !rest.includes("/")) {
      return { name: route.name, params: { [route.param]: decodeURIComponent(rest) } };
    }
  }
  return { name: "not-found", params: {} };
}

// 空 hash 规范成 #/，避免地址栏与断言里出现两种空值。
export function normalizeHash(): void {
  if (location.hash === "" || location.hash === "#") {
    history.replaceState(null, "", "#/");
  }
}

// 订阅路由变化，返回取消订阅的函数。
export function onRouteChange(handler: (route: ParsedRoute) => void): () => void {
  const listener = (): void => handler(parseRoute());
  window.addEventListener("hashchange", listener);
  return () => window.removeEventListener("hashchange", listener);
}
