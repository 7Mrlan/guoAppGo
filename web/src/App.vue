<script setup lang="ts">
// The M1-1 shell: enough markup for the acceptance checks to assert on, no
// product pages yet.
const status = "System 服务已绑定，页面尚未调用（M0）";
</script>

<template>
  <main>
    <h1>guoAppGo</h1>
    <p>桌面骨架已启动，前端由二进制内嵌提供。</p>
    <p><code data-role="bridge-status" data-bundle="ready">{{ status }}</code></p>
  </main>
</template>
