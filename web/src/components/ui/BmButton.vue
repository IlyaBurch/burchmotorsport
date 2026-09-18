<script setup lang="ts">
interface Props {
  variant?: 'default' | 'primary' | 'ghost'
  disabled?: boolean
  tag?: 'button' | 'a'
  type?: 'button' | 'submit' | 'reset'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  disabled: false,
  tag: 'button',
  type: 'button',
})

const classes = computed(() => [
  'bm-btn',
  props.variant !== 'default' && `bm-btn--${props.variant}`,
])

import { computed } from 'vue'
</script>

<template>
  <component
    :is="tag"
    :class="classes"
    :disabled="disabled || undefined"
    :type="tag === 'button' ? type : undefined"
  >
    <slot />
  </component>
</template>
