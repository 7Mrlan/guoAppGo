<script setup lang="ts">
// 通用按钮：primary 主操作、secondary 次操作、ghost 无底色。
withDefaults(
  defineProps<{
    variant?: "primary" | "secondary" | "ghost";
    disabled?: boolean;
  }>(),
  { variant: "secondary", disabled: false },
);
</script>

<template>
  <button type="button" class="btn" :class="`btn--${variant}`" :disabled="disabled">
    <slot />
  </button>
</template>
