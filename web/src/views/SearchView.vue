<script setup lang="ts">
// 搜索页：输入停下 300ms 后才查，避免每个字符都打一次宿主。
import { onBeforeUnmount, ref } from "vue";

import EmptyState from "../components/EmptyState.vue";
import TextField from "../components/TextField.vue";
import { host } from "../data/fixture";

const DEBOUNCE_MS = 300;

const keyword = ref("");
const results = ref<ReturnType<typeof host.search> | null>(null);
let timer: number | undefined;

// 输入即计时；清空立即回到初始态，不保留上一次结果。
function onInput(value: string): void {
  keyword.value = value;
  if (timer !== undefined) {
    window.clearTimeout(timer);
  }
  const query = value.trim();
  if (query === "") {
    results.value = null;
    timer = undefined;
    return;
  }
  timer = window.setTimeout(() => {
    results.value = host.search(query);
    timer = undefined;
  }, DEBOUNCE_MS);
}

onBeforeUnmount(() => {
  if (timer !== undefined) {
    window.clearTimeout(timer);
  }
});
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
      @input="onInput"
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
    <p v-else class="page__lead" data-role="search-idle">输入关键词后显示结果。</p>
  </section>
</template>
