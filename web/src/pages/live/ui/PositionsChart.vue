<script setup lang="ts">
import { computed } from 'vue'
import type { LiveDriver } from '@/shared/api/live'
import DriverPlate from './DriverPlate.vue'

const props = defineProps<{ drivers: LiveDriver[] }>()

const LAP_W = 28
const ROW_H = 18
const PAD = { l: 36, r: 12, t: 8, b: 24 }

const laps = computed(() => Math.max(0, ...props.drivers.map((d) => d.positions.length)))
const rows = computed(() => props.drivers.length)
const w = computed(() => PAD.l + laps.value * LAP_W + PAD.r)
const h = computed(() => PAD.t + rows.value * ROW_H + PAD.b)

const x = (lap: number) => PAD.l + (lap - 0.5) * LAP_W
const y = (pos: number) => PAD.t + (pos - 0.5) * ROW_H

// second car of a team is dashed so shared team colour still reads as two series
const dashed = computed(() => {
  const seen = new Set<string>()
  const out = new Map<number, boolean>()
  for (const d of props.drivers) {
    out.set(d.number, seen.has(d.team))
    seen.add(d.team)
  }
  return out
})

const series = computed(() =>
  props.drivers.map((d) => ({
    d,
    points: d.positions
      .map((p, i) => (p ? `${x(i + 1)},${y(p)}` : null))
      .filter(Boolean)
      .join(' '),
  })),
)

const lapTicks = computed(() => Array.from({ length: laps.value }, (_, i) => i + 1).filter((l) => l % 5 === 0 || l === 1))
</script>

<template>
  <div class="chart">
    <div class="chart__scroll">
      <svg :viewBox="`0 0 ${w} ${h}`" :style="{ minWidth: w + 'px' }" class="chart__svg" role="img" aria-label="Позиции по кругам">
        <g class="chart__grid">
          <line v-for="p in rows" :key="'r' + p" :x1="PAD.l" :x2="w - PAD.r" :y1="y(p)" :y2="y(p)" />
          <line v-for="l in lapTicks" :key="'l' + l" :x1="x(l)" :x2="x(l)" :y1="PAD.t" :y2="h - PAD.b" />
        </g>
        <g class="chart__axis timing-sm">
          <text v-for="p in rows" :key="'p' + p" :x="PAD.l - 6" :y="y(p)" text-anchor="end" dominant-baseline="central">P{{ p }}</text>
          <text v-for="l in lapTicks" :key="'t' + l" :x="x(l)" :y="h - 8" text-anchor="middle">{{ l }}</text>
        </g>
        <polyline
          v-for="s in series"
          :key="s.d.number"
          :points="s.points"
          fill="none"
          :stroke="s.d.teamColour ? '#' + s.d.teamColour : 'var(--text-muted)'"
          stroke-width="2"
          stroke-linejoin="round"
          :stroke-dasharray="dashed.get(s.d.number) ? '6 4' : undefined"
        >
          <title>{{ s.d.acronym }} · {{ s.d.name }}</title>
        </polyline>
      </svg>
    </div>
    <ul class="chart__legend">
      <li v-for="d in drivers" :key="d.number">
        <i :class="{ 'chart__swatch--dashed': dashed.get(d.number) }" :style="{ color: d.teamColour ? '#' + d.teamColour : 'var(--text-muted)' }" aria-hidden="true" />
        <DriverPlate :label="d.acronym" :colour="d.teamColour" />
      </li>
    </ul>
  </div>
</template>

<style scoped>
.chart {
  display: grid;
  gap: var(--space-3);
}

.chart__scroll {
  overflow-x: auto;
}

.chart__svg {
  display: block;
  width: 100%;
  height: auto;
}

.chart__grid line {
  stroke: var(--border);
  stroke-width: 1;
  stroke-dasharray: 1 3;
}

.chart__axis text {
  fill: var(--text-muted);
  font-family: var(--font-mono);
  font-size: 11px;
}

.chart__legend {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-2);
}

.chart__legend li {
  max-width: none; /* bm.css caps li at 65ch for prose */
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
}

.chart__legend i {
  width: 18px;
  height: 3px;
  background: currentColor;
}

.chart__swatch--dashed {
  height: 0;
  background: none;
  border-top: 3px dashed currentColor;
}
</style>
