<script setup lang="ts">
import { inkOn } from '@/shared/api/live'

defineProps<{ label: string; colour?: string; wide?: boolean }>()
</script>

<template>
  <!-- team colour is data from openf1, not a token: inline background, ink picked by luminance -->
  <span
    class="plate body-strong"
    :class="{ 'plate--wide': wide }"
    :style="colour ? { background: '#' + colour, color: inkOn(colour) } : undefined"
  >{{ label }}</span>
</template>

<style scoped>
.plate {
  display: inline-block;
  min-width: 48px;
  padding: 0 var(--space-2);
  border: var(--border-thin) solid var(--border);
  text-align: center;
  white-space: nowrap;
  background: var(--surface-raised);
  color: var(--text);
}

.plate--wide {
  display: block;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
