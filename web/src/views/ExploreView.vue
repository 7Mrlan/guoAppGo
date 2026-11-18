<script setup lang="ts">
// 发现页：分类切换 + 分页，目录走带模块版本的缓存。
import { ref } from "vue";

import EmptyState from "../components/EmptyState.vue";
import { readCatalog } from "../data/catalog";
import { host } from "../data/fixture";

const categories = host.categories().items;
const active = ref(categories[0]?.id ?? "");
const page = ref(1);
const result = ref(readCatalog(active.value, page.value));

// 换分类从第一页重新取；同分类重复进入由缓存挡住宿主调用。
function select(category: string): void {
  active.value = category;
  page.value = 1;
  result.value = readCatalog(category, 1);
}

// 有下一页就继续取，缓存键按页区分。
function more(): void {
  page.value += 1;
  result.value = readCatalog(active.value, page.value);
}

function first(): void {
  page.value = 1;
  result.value = readCatalog(active.value, 1);
}
</script>

<template>
  <section class="page">
    <h1 class="page__title">发现</h1>
    <p class="page__lead">浏览片库，点条目进入详情。</p>

    <div class="row">
      <button
        v-for="category in categories"
        :key="category.id"
        type="button"
        class="btn"
        :class="category.id === active ? 'btn--primary' : 'btn--secondary'"
        :aria-pressed="category.id === active"
        :data-category="category.id"
        @click="select(category.id)"
      >
        {{ category.name }}
      </button>
    </div>

    <ul v-if="result.items.length > 0" class="list">
      <li v-for="item in result.items" :key="item.id" class="list__row">
        <a class="list__item" :href="`#/detail/${item.id}`" data-role="catalog-item" :data-id="item.id">
          <span class="list__title">{{ item.title }}</span>
          <span class="list__meta">{{ item.subtitle }}</span>
        </a>
      </li>
    </ul>

    <EmptyState
      v-else
      :title="result.warning ?? '这里还没有内容'"
      hint="换个分类或稍后再来。"
    />

    <div class="row">
      <button v-if="page > 1" type="button" class="btn btn--secondary" data-role="page-prev" @click="first">
        回到第一页
      </button>
      <button v-if="result.hasMore" type="button" class="btn btn--primary" data-role="page-more" @click="more">
        下一页
      </button>
    </div>
  </section>
</template>
