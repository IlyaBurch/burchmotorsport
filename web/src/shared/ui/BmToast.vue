<script setup lang="ts">
interface Props {
  visible: boolean
}

defineProps<Props>()
</script>

<template>
  <Teleport to="body">
    <Transition name="toast">
      <div v-if="visible" class="bm-toast" role="status" aria-live="polite">
        <slot />
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: opacity 120ms ease-out;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
}
</style>
