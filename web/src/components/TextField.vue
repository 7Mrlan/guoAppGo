<script setup lang="ts">
// 单行输入框；根节点就是 input，属性（含 aria-label、data-ac）直接落在输入元素上。
withDefaults(
  defineProps<{
    type?: string;
    value?: string;
    placeholder?: string;
  }>(),
  { type: "text", value: "", placeholder: "" },
);

defineEmits<{ (e: "input", value: string): void }>();
</script>

<template>
  <input
    class="field"
    :type="type"
    :value="value"
    :placeholder="placeholder"
    @input="$emit('input', ($event.target as HTMLInputElement).value)"
  />
</template>
