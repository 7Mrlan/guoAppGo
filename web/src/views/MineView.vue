<script setup lang="ts">
// 我的：用户信息与各功能入口；未实现的入口显式不可用，不留死链。
const entries = [
  { label: "设置", href: "#/settings", ready: true },
  { label: "下载", href: "#/downloads", ready: false },
  { label: "收藏", href: "#/favorites", ready: false },
];
</script>

<template>
  <section class="page" data-role="mine">
    <h1 class="page__title">我的</h1>
    <p class="page__lead">本机用户，尚未接入账号。</p>

    <ul class="list">
      <li v-for="entry in entries" :key="entry.label" class="list__row">
        <a v-if="entry.ready" class="list__item" :href="entry.href" data-role="mine-entry">
          <span class="list__title">{{ entry.label }}</span>
        </a>
        <span v-else class="list__item" data-role="mine-entry-pending" aria-disabled="true">
          <span class="list__title">{{ entry.label }}</span>
          <span class="list__meta">尚未实现</span>
        </span>
      </li>
    </ul>
  </section>
</template>
