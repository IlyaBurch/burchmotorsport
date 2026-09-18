<script setup lang="ts">
import { Skull, Zap } from 'lucide-vue-next'
</script>

<template>
  <!-- decorative: wings of lightning around a skull, all lucide, no third-party art -->
  <span class="emblem" aria-hidden="true" title="Ave Dominus Nox">
    <Zap :size="20" :stroke-width="2.5" class="emblem__wing emblem__wing--l" />
    <Skull :size="24" :stroke-width="2.5" class="emblem__skull" />
    <Zap :size="20" :stroke-width="2.5" class="emblem__wing" />
  </span>
</template>

<style scoped>
.emblem {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  color: var(--accent); /* crimson wings */
}

.emblem__skull {
  color: var(--text); /* bone */
}

.emblem__wing--l {
  transform: scaleX(-1);
}
</style>
