<script setup lang="ts">
// 设置页：外观选项，跟随系统或固定某一主题。
import { ref } from "vue";

import { applyTheme, setThemePref, storedPref, type ThemePref } from "../theme";

const pref = ref<ThemePref>(storedPref());

const options: Array<{ value: ThemePref; label: string; hint: string }> = [
  { value: "system", label: "跟随系统", hint: "系统切换外观时页面同步" },
  { value: "light", label: "亮色", hint: "始终使用亮色主题" },
  { value: "dark", label: "暗色", hint: "始终使用暗色主题" },
];

// 立即生效并写入本地设置，下次启动沿用它。
function choose(value: ThemePref): void {
  pref.value = value;
  setThemePref(value);
}

// 挂载时应用一次已保存的偏好。
applyTheme(pref.value);
</script>

<template>
  <section class="page">
    <h1 class="page__title">设置</h1>
    <p class="page__lead">外观设置保存在本机，重开应用后仍然生效。</p>

    <div class="stack">
      <button
        v-for="option in options"
        :key="option.value"
        type="button"
        class="option"
        :aria-pressed="pref === option.value"
        @click="choose(option.value)"
      >
        <span class="option__label">{{ option.label }}</span>
        <span class="option__hint">{{ option.hint }}</span>
      </button>
    </div>
  </section>
</template>
