<script setup lang="ts">
interface Props {
  open: boolean
}

defineProps<Props>()

defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <template v-if="open">
      <div class="bm-overlay" @click="$emit('close')" />
      <div class="bm-modal" role="dialog" aria-modal="true">
        <slot />
      </div>
    </template>
  </Teleport>
</template>
