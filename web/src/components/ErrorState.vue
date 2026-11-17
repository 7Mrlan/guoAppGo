<script setup lang="ts">
// 错误状态：说明失败原因并提供重试入口。
withDefaults(
  defineProps<{
    title?: string;
    detail?: string;
  }>(),
  { title: "出错了", detail: "" },
);

defineEmits<{ (e: "retry"): void }>();
</script>

<template>
  <div class="state state--error" role="alert">
    <p class="state__title">{{ title }}</p>
    <p v-if="detail" class="state__hint">{{ detail }}</p>
    <button type="button" class="btn btn--secondary" @click="$emit('retry')">重试</button>
  </div>
</template>
