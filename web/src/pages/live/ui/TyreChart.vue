<script setup lang="ts">
import { computed } from 'vue'
import { compoundClass, type LiveDriver } from '@/shared/api/live'
import { DriverPlate } from '@/entities/session'

const props = defineProps<{ drivers: LiveDriver[] }>()

const total = computed(() => Math.max(1, ...props.drivers.flatMap((d) => (d.stints ?? []).map((s) => s.to))))
const pct = (n: number) => `${(n / total.value) * 100}%`

const LEGEND = [
  ['soft', 'Soft'],
  ['medium', 'Medium'],
  ['hard', 'Hard'],
  ['inter', 'Inter'],
  ['wet', 'Wet'],
] as const
</script>

<template>
  <div class="tyres">
    <ul class="tyres__legend">
      <li v-for="[cls, label] in LEGEND" :key="cls">
        <i :class="`tyres__seg--${cls}`" aria-hidden="true" /><span class="body-sm">{{ label }}</span>
      </li>
    </ul>
    <ol class="tyres__rows">
      <li v-for="d in drivers" :key="d.number" class="tyres__row">
        <DriverPlate :label="d.acronym" :colour="d.teamColour" />
        <div class="tyres__bar" role="img" :aria-label="`${d.acronym}: ${(d.stints ?? []).map((s) => `${s.compound} ${s.from}-${s.to}`).join(', ')}`">
          <span
            v-for="s in d.stints ?? []"
            :key="s.from"
            class="tyres__seg timing-sm"
            :class="`tyres__seg--${compoundClass(s.compound)}`"
            :style="{ left: pct(s.from - 1), width: pct(s.to - s.from + 1) }"
            :title="`${s.compound} · круги ${s.from}–${s.to}`"
          >{{ s.to - s.from + 1 }}</span>
        </div>
      </li>
    </ol>
    <div class="tyres__axis timing-sm">
      <span v-for="l in [1, Math.round(total / 4), Math.round(total / 2), Math.round((3 * total) / 4), total]" :key="l" :style="{ left: pct(l - 0.5) }">{{ l }}</span>
    </div>
  </div>
</template>

<style scoped>
.tyres {
  display: grid;
  gap: var(--space-3);
}

.tyres__legend,
.tyres__rows {
  list-style: none;
  margin: 0;
  padding: 0;
}

.tyres__legend {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
}

.tyres__legend li {
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
}

.tyres__legend i {
  width: 14px;
  height: 14px;
  border: 1px solid var(--border);
}

.tyres__rows {
  display: grid;
  gap: var(--space-1);
}

.tyres__row {
  max-width: none; /* bm.css caps li at 65ch for prose */
  display: grid;
  grid-template-columns: 56px 1fr;
  gap: var(--space-2);
  align-items: center;
}

.tyres__bar {
  position: relative;
  height: 20px;
  background: var(--surface-sunken);
  border: 1px solid var(--border);
}

.tyres__seg {
  position: absolute;
  top: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--ink);
  border-right: 2px solid var(--surface-raised); /* 2px surface gap between fills */
  overflow: hidden;
}

/* F1 compound convention, mapped to tokens */
.tyres__seg--soft { background: var(--flag-red); }
.tyres__seg--medium { background: var(--timing-yellow); }
.tyres__seg--hard { background: var(--paper); }
.tyres__seg--inter { background: var(--timing-green); }
.tyres__seg--wet { background: var(--accent); color: var(--on-accent); }
.tyres__seg--unknown { background: var(--text-muted); }

.tyres__axis {
  position: relative;
  height: 16px;
  margin-left: calc(56px + var(--space-2));
  color: var(--text-muted);
}

.tyres__axis span {
  position: absolute;
  transform: translateX(-50%);
}
</style>
