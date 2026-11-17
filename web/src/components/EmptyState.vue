<script setup lang="ts">
// 空状态：无数据时的说明与可选操作。
defineProps<{ title: string; hint?: string }>();
</script>

<template>
  <div class="state state--empty" role="status">
    <p class="state__title">{{ title }}</p>
    <p v-if="hint" class="state__hint">{{ hint }}</p>
    <slot />
  </div>
</template>
