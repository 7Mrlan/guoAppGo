<script setup lang="ts">
// 加载中：无确定进度，只占位不阻塞其它区域。
withDefaults(defineProps<{ label?: string }>(), { label: "加载中…" });
</script>

<template>
  <div class="state state--loading" role="status" aria-live="polite">
    <span class="state__spinner" aria-hidden="true" />
    <p class="state__title">{{ label }}</p>
  </div>
</template>
