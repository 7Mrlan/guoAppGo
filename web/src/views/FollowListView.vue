<script setup lang="ts">
// 追剧列表：按观看进度列出条目，点进去看单条的追剧详情。
import { computed } from "vue";

import EmptyState from "../components/EmptyState.vue";
import { host } from "../data/fixture";
import { recentProgress } from "../data/progress";

const entries = computed(() =>
  recentProgress().map((entry) => ({ ...entry, title: host.detail(entry.id)?.title ?? entry.id })),
);
</script>

<template>
  <section class="page" data-role="follow-list">
    <h1 class="page__title">追剧</h1>
    <p class="page__lead">正在追的条目与上次看到的位置。</p>

    <ul v-if="entries.length > 0" class="list">
      <li v-for="entry in entries" :key="entry.episodeId" class="list__row">
        <a class="list__item" :href="`#/follow/${entry.id}`" :data-id="entry.id" data-role="follow-item">
          <span class="list__title">{{ entry.title }}</span>
          <span class="list__meta">{{ entry.episodeId }}</span>
        </a>
      </li>
    </ul>
    <EmptyState v-else title="还没有在追的内容" hint="看过一集之后会出现在这里。" />
  </section>
</template>
