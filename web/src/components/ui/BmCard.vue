<script setup lang="ts">
interface Props {
  hoverable?: boolean
}

withDefaults(defineProps<Props>(), {
  hoverable: false,
})
</script>

<template>
  <div class="bm-card" :class="{ 'bm-card--hoverable': hoverable }">
    <slot />
  </div>
</template>

<style scoped>
.bm-card--hoverable {
  cursor: pointer;
  transition: transform 100ms ease-out, box-shadow 100ms ease-out;
}

.bm-card--hoverable:hover {
  transform: translate(-2px, -2px);
  box-shadow: var(--shadow-hover);
}

.bm-card--hoverable:active {
  transform: translate(4px, 4px);
  box-shadow: var(--shadow-pressed);
}
</style>
