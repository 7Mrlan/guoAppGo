<script setup lang="ts">
// 弹层：打开时焦点移入面板，Esc 或点遮罩关闭，关闭后焦点回到打开前的元素。
import { nextTick, onBeforeUnmount, ref, watch } from "vue";

const props = defineProps<{ open: boolean; title: string }>();
const emit = defineEmits<{ (e: "close"): void }>();

const panel = ref<HTMLElement | null>(null);
let returnFocus: HTMLElement | null = null;

// 弹层内可获得焦点的元素，用于 Tab 循环。
const FOCUSABLE =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';

// Tab / Shift+Tab 在面板内回绕；面板没有可聚焦元素时焦点留在面板上。
function trapTab(event: KeyboardEvent): void {
  const items = Array.from(panel.value?.querySelectorAll<HTMLElement>(FOCUSABLE) ?? []);
  if (items.length === 0) {
    event.preventDefault();
    panel.value?.focus();
    return;
  }
  const first = items[0];
  const last = items[items.length - 1];
  const active = document.activeElement as HTMLElement | null;
  const outside = !panel.value?.contains(active);
  if (event.shiftKey && (active === first || active === panel.value || outside)) {
    event.preventDefault();
    last.focus();
  } else if (!event.shiftKey && (active === last || outside)) {
    event.preventDefault();
    first.focus();
  }
}

// Esc 关闭；Tab 循环；监听挂在 document 上，焦点在弹层内任一控件时也能生效。
function onKeydown(event: KeyboardEvent): void {
  if (!props.open) {
    return;
  }
  if (event.key === "Escape") {
    event.preventDefault();
    emit("close");
  } else if (event.key === "Tab") {
    trapTab(event);
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
