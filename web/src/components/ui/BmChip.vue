<script setup lang="ts">
interface Props {
  variant?: 'default' | 'purple' | 'green' | 'yellow' | 'red'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
})

import { computed } from 'vue'

const classes = computed(() => [
  'bm-chip',
  props.variant !== 'default' && `bm-chip--${props.variant}`,
])
</script>

<template>
  <span :class="classes">
    <slot />
  </span>
</template>
