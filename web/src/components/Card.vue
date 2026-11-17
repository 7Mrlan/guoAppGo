<script setup lang="ts">
// 内容卡片：标题 + 默认插槽，可选点击。
withDefaults(
  defineProps<{
    title: string;
    interactive?: boolean;
  }>(),
  { interactive: false },
);
</script>

<template>
  <component :is="interactive ? 'button' : 'section'" class="card" :class="{ 'card--interactive': interactive }" :type="interactive ? 'button' : undefined">
    <h3 class="card__title">{{ title }}</h3>
    <div class="card__body"><slot /></div>
  </component>
</template>
