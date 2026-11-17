// hash 路由：路径与视图的对应关系，以及 hash 的规范化。
export type RouteName = "home" | "explore" | "settings" | "components" | "not-found";

const ROUTES: Record<string, RouteName> = {
  "": "home",
  "/": "home",
  "/explore": "explore",
  "/settings": "settings",
  "/dev/components": "components",
};

// 去掉查询串后查表，未知路径落到 not-found。
export function currentRoute(): RouteName {
  const path = location.hash.replace(/^#/, "").split("?")[0];
  return ROUTES[path] ?? "not-found";
}

// 空 hash 规范成 #/，避免地址栏与断言里出现两种空值。
export function normalizeHash(): void {
  if (location.hash === "" || location.hash === "#") {
    history.replaceState(null, "", "#/");
  }
}

// 订阅路由变化，返回取消订阅的函数。
export function onRouteChange(handler: (route: RouteName) => void): () => void {
  const listener = (): void => handler(currentRoute());
  window.addEventListener("hashchange", listener);
  return () => window.removeEventListener("hashchange", listener);
}
