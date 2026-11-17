<script setup lang="ts">
// 发现页：片单列表与空状态的组合。
import { ref } from "vue";

import EmptyState from "../components/EmptyState.vue";

const items = ref([
  { id: 1, title: "深海列车", meta: "剧情 · 2024 · 108 分钟" },
  { id: 2, title: "北纬三十度", meta: "纪录片 · 2023 · 6 集" },
  { id: 3, title: "夜行线路", meta: "悬疑 · 2025 · 12 集" },
]);
</script>

<template>
  <section class="page">
    <h1 class="page__title">发现</h1>
    <p class="page__lead">浏览片库，点条目进入详情。</p>

    <ul class="list">
      <li v-for="item in items" :key="item.id" class="list__row">
        <button type="button" class="list__item">
          <span class="list__title">{{ item.title }}</span>
          <span class="list__meta">{{ item.meta }}</span>
        </button>
      </li>
    </ul>

    <EmptyState v-if="items.length === 0" title="这里还没有内容" hint="换个分类或稍后再来。" />
  </section>
</template>
