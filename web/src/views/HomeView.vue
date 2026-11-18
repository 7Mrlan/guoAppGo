<script setup lang="ts">
// 首页：继续观看、为你推荐、榜单三个区块，各自处理空态。
import { computed, ref } from "vue";

import EmptyState from "../components/EmptyState.vue";
import ErrorState from "../components/ErrorState.vue";
import { host } from "../data/fixture";
import { recentProgress } from "../data/progress";

const error = ref("");
const feed = ref<ReturnType<typeof host.catalogPage>["items"]>([]);
const ranked = ref<ReturnType<typeof host.catalogPage>["items"]>([]);

// 宿主同步返回；抛错时进入错误态，重试即重新读取。
function load(): void {
  try {
    feed.value = host.catalogPage(1, "").items;
    ranked.value = host.catalogPage(1, "series").items.slice(0, 3);
    error.value = "";
  } catch (err) {
    feed.value = [];
    ranked.value = [];
    error.value = err instanceof Error ? err.message : String(err);
  }
}

load();

const watching = computed(() => recentProgress());
</script>

<template>
  <section class="page" data-role="home">
    <h1 class="page__title">guoAppGo</h1>
    <p class="page__lead">桌面骨架已启动，内容由本地数据提供。</p>

    <ErrorState v-if="error" data-role="home-error" :detail="error" @retry="load" />

    <template v-else>
      <h2 class="section__title">继续观看</h2>
      <ul v-if="watching.length > 0" class="list" data-role="home-watching">
        <li v-for="entry in watching" :key="entry.episodeId" class="list__row">
          <a class="list__item" :href="`#/detail/${entry.id}`" :data-id="entry.id" data-role="home-watching-item">
            <span class="list__title">{{ host.detail(entry.id)?.title ?? entry.id }}</span>
            <span class="list__meta">{{ entry.episodeId }} · 已看 {{ Math.round(entry.positionMs / 1000) }} 秒</span>
          </a>
        </li>
      </ul>
      <EmptyState v-else data-role="home-watching-empty" title="还没有观看记录" hint="从发现页挑一部开始。" />

      <h2 class="section__title">为你推荐</h2>
      <ul v-if="feed.length > 0" class="list" data-role="home-feed">
        <li v-for="item in feed" :key="item.id" class="list__row">
          <a class="list__item" :href="`#/detail/${item.id}`" :data-id="item.id">
            <span class="list__title">{{ item.title }}</span>
            <span class="list__meta">{{ item.subtitle }}</span>
          </a>
        </li>
      </ul>
      <EmptyState v-else data-role="home-feed-empty" title="暂无推荐" hint="稍后再来。" />

      <h2 class="section__title">榜单</h2>
      <ul v-if="ranked.length > 0" class="list" data-role="home-ranking">
        <li v-for="(item, index) in ranked" :key="item.id" class="list__row">
          <a class="list__item" :href="`#/detail/${item.id}`" :data-id="item.id">
            <span class="list__title">{{ index + 1 }}. {{ item.title }}</span>
            <span class="list__meta">{{ item.subtitle }}</span>
          </a>
        </li>
      </ul>
      <EmptyState v-else data-role="home-ranking-empty" title="暂无榜单" hint="稍后再来。" />
    </template>
  </section>
</template>
