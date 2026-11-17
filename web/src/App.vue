<script setup lang="ts">
// 应用外壳：标题、启动说明与 bridge 状态占位。
const status = "System 服务已绑定，页面尚未调用（M0）";
</script>

<template>
  <main>
    <h1>guoAppGo</h1>
    <p>桌面骨架已启动，前端由二进制内嵌提供。</p>
    <p><code data-role="bridge-status" data-bundle="ready">{{ status }}</code></p>
  </main>
</template>
