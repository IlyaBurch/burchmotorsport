<script setup lang="ts">
import { computed } from 'vue'
import type { LiveDriver } from '@/shared/api/live'

const props = defineProps<{
  outline: [number, number][]
  drivers: LiveDriver[]
}>()

// openf1 y grows "up", svg y grows down → flip once here
const box = computed(() => {
  const xs = props.outline.map((p) => p[0])
  const ys = props.outline.map((p) => p[1])
  const minX = Math.min(...xs)
  const maxX = Math.max(...xs)
  const minY = Math.min(...ys)
  const maxY = Math.max(...ys)
  const pad = Math.max(maxX - minX, maxY - minY) * 0.08
  return { minX: minX - pad, maxY: maxY + pad, w: maxX - minX + pad * 2, h: maxY - minY + pad * 2 }
})

const flipY = (y: number) => box.value.maxY - y

const path = computed(() =>
  props.outline.map((p, i) => `${i ? 'L' : 'M'}${p[0]} ${flipY(p[1])}`).join(' ') + ' Z',
)

const cars = computed(() => props.drivers.filter((d) => d.x || d.y))

// stroke and marker sizes in track units, so they scale with the circuit
const unit = computed(() => box.value.w / 100)
</script>

<template>
  <svg
    class="map"
    :viewBox="`${box.minX} 0 ${box.w} ${box.h}`"
    role="img"
    aria-label="Положение пилотов на трассе"
  >
    <path :d="path" fill="none" stroke="var(--border)" :stroke-width="unit * 2.2" stroke-linejoin="round" />
    <path :d="path" fill="none" stroke="var(--surface-sunken)" :stroke-width="unit * 1.2" stroke-linejoin="round" />
    <g v-for="d in cars" :key="d.number" :transform="`translate(${d.x} ${flipY(d.y)})`">
      <circle :r="unit * 2.4" :fill="d.position === 1 ? 'var(--accent)' : 'var(--surface-raised)'" stroke="var(--border)" :stroke-width="unit * 0.5" />
      <text
        text-anchor="middle"
        dominant-baseline="central"
        :font-size="unit * 2.2"
        :fill="d.position === 1 ? 'var(--on-accent)' : 'var(--text)'"
        class="map__label"
      >
        {{ d.acronym }}
      </text>
    </g>
  </svg>
</template>

<style scoped>
.map {
  display: block;
  width: 100%;
  height: auto;
  background: var(--surface-raised);
  border: var(--border-thick) solid var(--border);
  box-shadow: var(--shadow-hard);
}

.map__label {
  font-family: var(--font-mono);
  font-weight: 600;
  pointer-events: none;
}
</style>
