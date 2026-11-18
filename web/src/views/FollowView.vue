<script setup lang="ts">
// 追剧页：单个条目的追剧状态与回到详情的入口。
import { computed } from "vue";

import { host } from "../data/fixture";

const props = defineProps<{ id: string }>();

const detail = computed(() => host.detail(props.id));
const lastEpisode = computed(() => detail.value?.seasons[0]?.episodes[0]);
</script>

<template>
  <section v-if="detail" class="page" data-role="follow">
    <h1 class="page__title" data-role="follow-title">{{ detail.title }}</h1>
    <p class="page__lead">追剧状态与上次观看位置会在这里显示。</p>

    <ul class="list">
      <li class="list__row">
        <span class="list__item">
          <span class="list__title">上次观看</span>
          <span class="list__meta">{{ lastEpisode?.title ?? "尚无记录" }}</span>
        </span>
      </li>
      <li class="list__row">
        <a class="list__item" :href="`#/detail/${detail.id}`">
          <span class="list__title">回到详情</span>
          <span class="list__meta">{{ detail.id }}</span>
        </a>
      </li>
    </ul>
  </section>
</template>
