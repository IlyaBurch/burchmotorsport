<script setup lang="ts">
interface Props {
  label?: string
  error?: string
  placeholder?: string
  type?: string
  modelValue?: string
}

withDefaults(defineProps<Props>(), {
  type: 'text',
  modelValue: '',
})

defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="bm-input-group">
    <label v-if="label" class="bm-input-group__label">{{ label }}</label>
    <input
      class="bm-input"
      :type="type"
      :placeholder="placeholder"
      :value="modelValue"
      @input="$emit('update:modelValue', ($event.target as HTMLInputElement).value)"
    />
    <BmChip v-if="error" variant="red">{{ error }}</BmChip>
  </div>
</template>

<script lang="ts">
import BmChip from './BmChip.vue'
</script>
