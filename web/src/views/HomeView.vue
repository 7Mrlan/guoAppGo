<script setup lang="ts">
// 首页：推荐卡片与启动状态说明。
import Card from "../components/Card.vue";

const status = "System 服务已绑定，页面尚未调用";
</script>

<template>
  <section class="page">
    <h1 class="page__title">guoAppGo</h1>
    <p class="page__lead">桌面骨架已启动，前端由二进制内嵌提供。</p>

    <div class="grid">
      <Card title="继续观看">上次播放记录会出现在这里。</Card>
      <Card title="为你推荐">按观看偏好排序的片单。</Card>
      <Card title="最近添加">片库最近入库的条目。</Card>
    </div>

    <p><code data-role="bridge-status" data-bundle="ready">{{ status }}</code></p>
  </section>
</template>
