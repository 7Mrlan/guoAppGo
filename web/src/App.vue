<script setup lang="ts">
// 应用外壳：主导航与当前路由对应的页面。
import { onBeforeUnmount, ref } from "vue";

import { currentRoute, normalizeHash, onRouteChange, type RouteName } from "./router";
import ComponentsView from "./views/ComponentsView.vue";
import ExploreView from "./views/ExploreView.vue";
import HomeView from "./views/HomeView.vue";
import NotFoundView from "./views/NotFoundView.vue";
import SettingsView from "./views/SettingsView.vue";

normalizeHash();

const route = ref<RouteName>(currentRoute());
// 取消路由订阅的函数，组件卸载时调用。
const stop = onRouteChange((next) => {
  route.value = next;
});
onBeforeUnmount(stop);

const NAV: Array<{ route: RouteName; href: string; label: string }> = [
  { route: "home", href: "#/", label: "首页" },
  { route: "explore", href: "#/explore", label: "发现" },
  { route: "settings", href: "#/settings", label: "设置" },
  { route: "components", href: "#/dev/components", label: "组件" },
];
</script>

<template>
  <header class="topbar">
    <span class="brand">guoAppGo</span>
    <nav class="nav" aria-label="主导航">
      <a
        v-for="item in NAV"
        :key="item.route"
        class="nav__link"
        :href="item.href"
        :data-route="item.route"
        :aria-current="route === item.route ? 'page' : undefined"
      >
        {{ item.label }}
      </a>
    </nav>
  </header>

  <main class="main">
    <HomeView v-if="route === 'home'" />
    <ExploreView v-else-if="route === 'explore'" />
    <SettingsView v-else-if="route === 'settings'" />
    <ComponentsView v-else-if="route === 'components'" />
    <NotFoundView v-else />
  </main>
</template>
