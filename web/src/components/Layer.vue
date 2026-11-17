<script setup lang="ts">
// 弹层：打开时焦点移入面板，Esc 或点遮罩关闭，关闭后焦点回到打开前的元素。
import { nextTick, onBeforeUnmount, ref, watch } from "vue";

const props = defineProps<{ open: boolean; title: string }>();
const emit = defineEmits<{ (e: "close"): void }>();

const panel = ref<HTMLElement | null>(null);
let returnFocus: HTMLElement | null = null;

// Esc 关闭；监听挂在 document 上，焦点在弹层内任一控件时也能生效。
function onKeydown(event: KeyboardEvent): void {
  if (event.key === "Escape" && props.open) {
    event.preventDefault();
    emit("close");
  }
}

watch(
  () => props.open,
  async (open) => {
    if (open) {
      returnFocus = document.activeElement as HTMLElement | null;
      document.addEventListener("keydown", onKeydown);
      await nextTick();
      panel.value?.focus();
    } else {
      document.removeEventListener("keydown", onKeydown);
      returnFocus?.focus();
      returnFocus = null;
    }
  },
);

onBeforeUnmount(() => document.removeEventListener("keydown", onKeydown));
</script>

<template>
  <div v-if="open" data-role="layer" class="layer" role="dialog" aria-modal="true" :aria-label="title" @click.self="emit('close')">
    <div ref="panel" class="layer__panel" tabindex="-1">
      <h2 class="layer__title">{{ title }}</h2>
      <slot />
    </div>
  </div>
</template>
