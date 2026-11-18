<script setup lang="ts">
// 搜索页：输入后按宿主搜索接口取结果，无结果给空状态。
import { ref } from "vue";

import EmptyState from "../components/EmptyState.vue";
import TextField from "../components/TextField.vue";
import { host } from "../data/fixture";

const keyword = ref("");
const results = ref<ReturnType<typeof host.search> | null>(null);

// 宿主搜索同步返回，页面不做防抖排队；输入即查，先满足可断言的行为。
function run(value: string): void {
  keyword.value = value;
  results.value = value.trim() === "" ? null : host.search(value);
}
</script>

<template>
  <section class="page" data-role="search">
    <h1 class="page__title">搜索</h1>
    <p class="page__lead">按片名搜索本机片单。</p>

    <TextField
      data-role="search-input"
      aria-label="搜索片名"
      placeholder="输入片名"
      :value="keyword"
      @input="run"
    />

    <ul v-if="results && results.items.length > 0" class="list" data-role="search-results">
      <li v-for="item in results.items" :key="item.id" class="list__row">
        <a class="list__item" :href="`#/detail/${item.id}`" :data-id="item.id">
          <span class="list__title">{{ item.title }}</span>
          <span class="list__meta">{{ item.subtitle }}</span>
        </a>
      </li>
    </ul>
    <EmptyState
      v-else-if="results"
      data-role="search-empty"
      :title="results.warning ?? '没有匹配的结果'"
      hint="换个关键词试试。"
    />
    <p v-else class="page__lead">输入关键词后显示结果。</p>
  </section>
</template>
