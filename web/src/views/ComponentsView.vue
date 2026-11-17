<script setup lang="ts">
// 组件预览页：按钮、输入、卡片、列表与空/加载/错误状态的实物。
import { ref } from "vue";

import Button from "../components/Button.vue";
import Card from "../components/Card.vue";
import EmptyState from "../components/EmptyState.vue";
import ErrorState from "../components/ErrorState.vue";
import Layer from "../components/Layer.vue";
import LoadingState from "../components/LoadingState.vue";
import TextField from "../components/TextField.vue";

const layerOpen = ref(false);
const query = ref("");
const sort = ref("recent");
const autoplay = ref(true);

// 打开弹层；焦点落在面板上，关闭后回到触发按钮，均由 Layer 处理。
function openLayer(): void {
  layerOpen.value = true;
}
</script>

<template>
  <section class="page">
    <h1 class="page__title">通用组件</h1>
    <p class="page__lead">按钮、输入、卡片、列表与各类状态。</p>

    <div class="stack">
      <div class="row">
        <Button data-ac="btn-primary" variant="primary">开始播放</Button>
        <Button data-ac="btn-secondary">加入收藏</Button>
        <Button data-ac="btn-ghost" variant="ghost">稍后再说</Button>
        <Button data-ac="btn-disabled" disabled>暂不可用</Button>
      </div>

      <div class="row">
        <TextField
          data-ac="field-search"
          aria-label="搜索关键词"
          placeholder="搜索片名"
          :value="query"
          @input="query = $event"
        />
        <TextField data-ac="field-url" type="url" aria-label="播放地址" placeholder="https://…" />
        <select v-model="sort" data-ac="select-sort" class="field" aria-label="排序方式">
          <option value="recent">最近添加</option>
          <option value="title">按片名</option>
          <option value="score">按评分</option>
        </select>
        <button
          type="button"
          data-ac="switch-autoplay"
          class="btn btn--secondary"
          role="switch"
          :aria-checked="autoplay"
          aria-label="自动播放"
          @click="autoplay = !autoplay"
        >
          自动播放
        </button>
      </div>

      <div class="row">
        <Card data-ac="card-interactive" title="卡片（可点击）" interactive>点开查看详情。</Card>
        <button type="button" data-ac="list-item-1" class="list__item">片单条目一</button>
        <button type="button" data-ac="list-item-2" class="list__item">片单条目二</button>
        <Button data-ac="layer-trigger" variant="primary" @click="openLayer">打开弹层</Button>
      </div>

      <EmptyState title="空状态" hint="没有满足条件的内容。" />
      <LoadingState />
      <ErrorState detail="网络请求失败，请检查连接。" />
    </div>

    <Layer :open="layerOpen" title="弹层示例" @close="layerOpen = false">
      <p>按 Esc 或点击遮罩关闭，焦点会回到打开它的按钮。</p>
    </Layer>
  </section>
</template>
