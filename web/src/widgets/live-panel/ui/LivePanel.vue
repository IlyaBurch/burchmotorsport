<script setup lang="ts">
import { computed, ref } from 'vue'
import { BmChip, BmTabs } from '@/shared/ui'
import { TrackMap, DriverPlate } from '@/entities/session'
import { compoundLetter, flagClass, formatGap, projectStandings, type Live } from '@/shared/api/live'

const props = defineProps<{
  live: Live | null
  outline: [number, number][]
  /** over-video variant: tighter, one tab at a time, no chrome */
  compact?: boolean
}>()

const tabs = [
  { key: 'track', label: 'Трасса' },
  { key: 'tyres', label: 'Шины' },
  { key: 'standings', label: 'Зачёт' },
]
const tab = ref('track')

const byNumber = computed(() => new Map(props.live?.drivers.map((d) => [d.number, d]) ?? []))
const standings = computed(() =>
  props.live
    ? [...projectStandings(props.live.session.session_name, props.live.drivers).values()]
        .sort((a, b) => a.pos - b.pos)
        .slice(0, 10)
    : [],
)
const raceControl = computed(() => props.live?.raceControl.slice(0, props.compact ? 4 : 8) ?? [])
</script>

<template>
  <section class="panel" :class="{ 'panel--compact': compact }">
    <BmTabs v-model="tab" :tabs="tabs" class="panel__tabs" />

    <div v-if="tab === 'track'" class="panel__body">
      <TrackMap v-if="outline.length && live" :outline="outline" :drivers="live.drivers" />
      <ol v-if="raceControl.length" class="panel__rc">
        <li v-for="m in raceControl" :key="m.date + m.message" class="panel__msg">
          <span class="timing-sm">L{{ m.lap_number || '—' }}</span>
          <i v-if="flagClass(m.flag)" class="panel__sq" :class="`panel__sq--${flagClass(m.flag)}`" aria-hidden="true" />
          <span class="body-sm">{{ m.message }}</span>
        </li>
      </ol>
      <p v-else-if="live" class="body-sm">Race control молчит</p>
    </div>

    <ol v-else-if="tab === 'tyres'" class="panel__body panel__rows">
      <li v-for="d in live?.drivers ?? []" :key="d.number" class="panel__row">
        <span class="display-sm">{{ d.position || '—' }}</span>
        <DriverPlate :label="d.acronym" :colour="d.teamColour" />
        <span class="timing">{{ d.compound ? `${compoundLetter(d.compound)} ${d.tyreAge}` : '—' }}</span>
        <span class="timing-sm panel__muted">{{ d.pits }} pit</span>
        <span class="timing-sm panel__num">{{ formatGap(d.interval) }}</span>
      </li>
    </ol>

    <ol v-else class="panel__body panel__rows">
      <li v-for="s in standings" :key="s.number" class="panel__row">
        <span class="display-sm">{{ s.pos }}</span>
        <DriverPlate :label="byNumber.get(s.number)?.acronym ?? ''" :colour="byNumber.get(s.number)?.teamColour" />
        <span class="timing">{{ s.points }}</span>
        <span class="timing-sm panel__muted">{{ s.gain ? `+${s.gain}` : '' }}</span>
        <BmChip v-if="s.delta" :variant="s.delta > 0 ? 'green' : 'yellow'">{{ s.delta > 0 ? `▲${s.delta}` : `▼${-s.delta}` }}</BmChip>
        <span v-else class="timing-sm panel__num">—</span>
      </li>
    </ol>
  </section>
</template>

<style scoped>
.panel {
  display: grid;
  gap: var(--space-3);
  align-content: start;
  min-width: 0;
}

.panel__body {
  display: grid;
  gap: var(--space-3);
  min-width: 0;
}

.panel__rc,
.panel__rows {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.panel__msg {
  max-width: none;
  display: grid;
  grid-template-columns: 40px auto 1fr;
  gap: var(--space-2);
  align-items: baseline;
}

.panel__row {
  max-width: none;
  display: grid;
  grid-template-columns: 28px 56px 1fr auto auto;
  gap: var(--space-2);
  align-items: center;
}

.panel__num {
  text-align: right;
}

.panel__muted {
  color: var(--text-muted);
}

.panel__sq {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 1px solid var(--border);
  align-self: center;
}

.panel__sq--yellow { background: var(--timing-yellow); }
.panel__sq--green { background: var(--timing-green); }
.panel__sq--red { background: var(--flag-red); }
.panel__sq--blue { background: var(--accent); }
.panel__sq--plain { background: var(--paper); }
.panel__sq--chequered { background: var(--ink); }

/* over the video: solid plate, scrolls inside, nothing bleeds out */
.panel--compact {
  padding: var(--space-3);
  background: var(--surface-raised);
  border: var(--border-thin) solid var(--border);
  box-shadow: var(--shadow-hard);
  max-height: 100%;
  overflow-y: auto;
}

.panel--compact .panel__msg .body-sm {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
