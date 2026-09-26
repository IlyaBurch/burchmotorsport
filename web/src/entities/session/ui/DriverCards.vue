<script setup lang="ts">
import { computed } from 'vue'
import { BmChip } from '@/shared/ui'
import { compoundLetter, formatGap, formatLap, formatSector, inkOn, type Live } from '@/shared/api/live'

type Driver = Live['drivers'][number]
const props = defineProps<{ drivers: Driver[] }>()

const min = (xs: (number | null | undefined)[]) => {
  const v = xs.filter((x): x is number => x != null)
  return v.length ? Math.min(...v) : null
}
const overallBestLap = computed(() => min(props.drivers.map((d) => d.bestLap)))
const overallBestSectors = computed(() => [0, 1, 2].map((i) => min(props.drivers.map((d) => d.bestSectors[i]))))

type Tone = 'purple' | 'green' | 'default'
const sectorTone = (d: Driver, i: number): Tone => {
  const s = d.sectors[i]
  if (s == null) return 'default'
  if (s === overallBestSectors.value[i]) return 'purple'
  if (s === d.bestSectors[i]) return 'green'
  return 'default'
}
const bestTone = (d: Driver, i: number): Tone =>
  d.bestSectors[i] != null && d.bestSectors[i] === overallBestSectors.value[i] ? 'purple' : 'default'
</script>

<template>
  <!-- one card per driver instead of a 12-column scrolling table -->
  <ol class="cards" aria-label="Пилоты">
    <li v-for="d in drivers" :key="d.number" class="card">
      <div class="card-id" :style="d.teamColour ? { background: '#' + d.teamColour, color: inkOn(d.teamColour) } : undefined">
        <span class="display-md">{{ d.position || '—' }}</span>
        <span class="body-strong">{{ d.acronym }}</span>
      </div>
      <div class="card-body">
        <div class="cell">
          <span class="timing">{{ formatGap(d.interval) }}</span>
          <span class="timing-sm muted">{{ formatGap(d.gap) }}</span>
        </div>
        <div class="cell">
          <span class="timing">{{ d.compound ? `${compoundLetter(d.compound)} ${d.tyreAge}` : '—' }}</span>
          <span class="timing-sm muted">{{ d.pits }} pit</span>
        </div>
        <div class="cell">
          <BmChip v-if="d.bestLap != null && d.bestLap === overallBestLap" variant="purple">{{ formatLap(d.lastLap) }}</BmChip>
          <span v-else class="timing">{{ formatLap(d.lastLap) }}</span>
          <span class="timing-sm muted">{{ formatLap(d.bestLap) }}</span>
        </div>
        <div class="cell cell--laps">
          <span class="timing">{{ d.lap }}</span>
          <span class="timing-sm muted">laps</span>
        </div>
        <div v-for="i in [0, 1, 2]" :key="i" class="cell cell--sector">
          <BmChip v-if="sectorTone(d, i) !== 'default'" :variant="sectorTone(d, i)">{{ formatSector(d.sectors[i]) }}</BmChip>
          <span v-else class="timing-sm">{{ formatSector(d.sectors[i]) }}</span>
          <BmChip v-if="bestTone(d, i) !== 'default'" :variant="bestTone(d, i)">{{ formatSector(d.bestSectors[i]) }}</BmChip>
          <span v-else class="timing-sm muted">{{ formatSector(d.bestSectors[i]) }}</span>
        </div>
      </div>
    </li>
  </ol>
</template>

<style scoped>
.cards {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.card {
  max-width: none;
  display: grid;
  grid-template-columns: 64px 1fr;
  background: var(--surface-sunken);
  border: var(--border-thin) solid var(--border);
}

.card-id {
  display: grid;
  place-content: center;
  text-align: center;
  background: var(--surface-raised);
  border-right: var(--border-thin) solid var(--border);
}

.card-body {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1.3fr 0.6fr;
  gap: 1px;
  background: var(--border);
}

.cell {
  display: grid;
  align-content: center;
  justify-items: center;
  gap: 2px;
  padding: var(--space-1);
  background: var(--surface-sunken);
  text-align: center;
}

.cell--laps {
  background: var(--surface-raised);
}

.card-body .cell--sector:nth-of-type(5) {
  grid-column: 1 / 2;
}

.card-body .cell--sector:nth-of-type(7) {
  grid-column: 3 / 5;
}

.muted {
  color: var(--text-muted);
}
</style>
