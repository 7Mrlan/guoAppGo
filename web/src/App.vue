<script setup lang="ts">
// 应用外壳：侧栏导航与当前路由对应的页面；窄窗口下侧栏折成横向导航。
import { computed, onBeforeUnmount, ref } from "vue";

import { host } from "./data/fixture";
import { normalizeHash, onRouteChange, parseRoute, type ParsedRoute } from "./router";
import ComponentsView from "./views/ComponentsView.vue";
import DetailView from "./views/DetailView.vue";
import ExploreView from "./views/ExploreView.vue";
import FollowListView from "./views/FollowListView.vue";
import FollowView from "./views/FollowView.vue";
import HomeView from "./views/HomeView.vue";
import MineView from "./views/MineView.vue";
import NotFoundView from "./views/NotFoundView.vue";
import SearchView from "./views/SearchView.vue";
import SettingsView from "./views/SettingsView.vue";

normalizeHash();

const route = ref<ParsedRoute>(parseRoute());
const stop = onRouteChange((next) => {
  route.value = next;
});
onBeforeUnmount(stop);

// 带参数的路由要求条目真实存在，否则按未找到处理。
const requestedId = computed(() => route.value.params.id ?? "");
const known = computed(() => requestedId.value !== "" && host.detail(requestedId.value) !== null);

const NAV = [
  { route: "home", href: "#/", label: "首页" },
  { route: "explore", href: "#/explore", label: "发现" },
  { route: "follow-list", href: "#/follow", label: "追剧" },
  { route: "mine", href: "#/mine", label: "我的" },
] as const;

// 当前路由是否属于该导航项；带参数的详情从发现进入时仍高亮发现。
function isCurrent(name: string): boolean {
  if (name === "explore") {
    return route.value.name === "explore" || route.value.name === "detail";
  }
  if (name === "follow-list") {
    return route.value.name === "follow-list" || route.value.name === "follow";
  }
  return route.value.name === name;
}
</script>

<template>
  <div class="shell">
    <aside class="sidebar" data-role="sidebar">
      <span class="brand">guoAppGo</span>
      <nav class="nav" aria-label="主导航">
        <a
          v-for="item in NAV"
          :key="item.route"
          class="nav__link"
          :href="item.href"
          :data-route="item.route"
          data-role="nav-link"
          :aria-current="isCurrent(item.route) ? 'page' : undefined"
        >
          {{ item.label }}
        </a>
        <a class="nav__link" href="#/search" data-route="search" :aria-current="route.name === 'search' ? 'page' : undefined">
          搜索
        </a>
      </nav>
    </aside>

    <div class="content">
      <main class="main">
        <HomeView v-if="route.name === 'home'" />
        <ExploreView v-else-if="route.name === 'explore'" />
        <SearchView v-else-if="route.name === 'search'" />
        <SettingsView v-else-if="route.name === 'settings'" />
        <ComponentsView v-else-if="route.name === 'components'" />
        <FollowListView v-else-if="route.name === 'follow-list'" />
        <MineView v-else-if="route.name === 'mine'" />
        <DetailView v-else-if="route.name === 'detail' && known" :id="requestedId" />
        <FollowView v-else-if="route.name === 'follow' && known" :id="requestedId" />
        <NotFoundView v-else />
      </main>
    </div>
  </div>
</template>
