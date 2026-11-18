<script setup lang="ts">
// 详情页：片名、可展开的简介、季切换与集列表；上次看到的那集带标记。
import { computed, ref } from "vue";

import { host } from "../data/fixture";
import { recentProgress } from "../data/progress";

const props = defineProps<{ id: string }>();

// 未知 id 交给外层渲染未找到，这里只负责取数据。
const detail = computed(() => host.detail(props.id));
const expanded = ref(false);
const season = ref(0);

const seasons = computed(() => detail.value?.seasons ?? []);
const episodes = computed(() => seasons.value[season.value]?.episodes ?? []);

// 观看进度里的那集在本片详情中用于标记"上次看到"。
const lastEpisode = computed(() => recentProgress().find((entry) => entry.id === props.id)?.episodeId ?? "");
</script>

<template>
  <section v-if="detail" class="page" data-role="detail">
    <h1 class="page__title" data-role="detail-title">{{ detail.title }}</h1>
    <p class="page__lead">
      <span data-role="detail-description" :class="{ detail__clamp: !expanded }">{{ detail.description }}</span>
      <button type="button" class="btn btn--ghost" data-role="detail-expand" @click="expanded = !expanded">
        {{ expanded ? "收起" : "展开全文" }}
      </button>
    </p>

    <div class="row">
      <button
        v-for="(item, index) in seasons"
        :key="item.id"
        type="button"
        class="btn"
        :class="index === season ? 'btn--primary' : 'btn--secondary'"
        :aria-pressed="index === season"
        :data-season="index"
        @click="season = index"
      >
        {{ item.title }}
      </button>
    </div>

    <ul class="list">
      <li v-for="episode in episodes" :key="episode.id" class="list__row">
        <a
          class="list__item"
          :href="`#/follow/${detail.id}`"
          :data-episode="episode.id"
          :data-current-episode="episode.id === lastEpisode ? 'true' : undefined"
        >
          <span class="list__title">{{ episode.title }}</span>
          <span class="list__meta">{{ episode.id === lastEpisode ? "上次看到这里" : episode.id }}</span>
        </a>
      </li>
    </ul>
  </section>
</template>
