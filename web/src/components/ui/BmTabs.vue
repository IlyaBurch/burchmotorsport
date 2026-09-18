<script setup lang="ts">
interface Tab {
  key: string
  label: string
}

interface Props {
  tabs: Tab[]
  modelValue: string
}

defineProps<Props>()

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="bm-tabs" role="tablist">
    <button
      v-for="tab in tabs"
      :key="tab.key"
      role="tab"
      class="bm-tabs__item"
      :aria-selected="modelValue === tab.key"
      @click="$emit('update:modelValue', tab.key)"
    >
      {{ tab.label }}
    </button>
  </div>
</template>
