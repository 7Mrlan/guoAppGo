<script setup lang="ts">
// 发现页：分类切换 + 片单列表，条目进详情。
import { ref } from "vue";

import EmptyState from "../components/EmptyState.vue";
import { host } from "../data/fixture";

const categories = host.categories().items;
const active = ref(categories[0]?.id ?? "");
const page = ref(host.catalogPage(1, active.value));

// 换分类就重取第一页；数据来自 api:1 宿主，接口同步返回。
function select(category: string): void {
  active.value = category;
  page.value = host.catalogPage(1, category);
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
        @click="select(category.id)"
      >
        {{ category.name }}
      </button>
    </div>

    <ul v-if="page.items.length > 0" class="list">
      <li v-for="item in page.items" :key="item.id" class="list__row">
        <a class="list__item" :href="`#/detail/${item.id}`" data-role="catalog-item" :data-id="item.id">
          <span class="list__title">{{ item.title }}</span>
          <span class="list__meta">{{ item.subtitle }}</span>
        </a>
      </li>
    </ul>

    <EmptyState
      v-else
      :title="page.warning ?? '这里还没有内容'"
      hint="换个分类或稍后再来。"
    />
  </section>
</template>
