<script setup lang="ts">
import { computed, ref } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { BmCard, BmChip } from '@/shared/ui'
import { fetchLive, formatGap, formatLap, type Live } from '@/shared/api/live'

const live = ref<Live | null>(null)
const error = ref<string | null>(null)

async function refresh() {
  try {
    live.value = await fetchLive()
    error.value = null
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
}

refresh()
useIntervalFn(refresh, 5000)

const bestLapNumber = computed(() => {
  const d = live.value?.drivers.filter((x) => x.bestLap != null) ?? []
  return d.length ? d.reduce((a, b) => (a.bestLap! <= b.bestLap! ? a : b)).number : null
})

const leaderLap = computed(() => live.value?.drivers[0]?.lap ?? 0)

const updated = computed(() =>
  live.value
    ? new Date(live.value.updatedAt).toLocaleTimeString('ru-RU', { timeZone: 'Europe/Moscow' }) + ' МСК'
    : '',
)
</script>

<template>
  <section class="live">
    <BmCard class="live__head">
      <div class="bm-card__eyebrow">Live</div>
      <h1 class="bm-card__title">
        {{ live ? `${live.session.circuit_short_name} · ${live.session.session_name}` : 'Телеметрия' }}
      </h1>
      <p class="body-sm live__meta">
        <template v-if="live">Круг {{ leaderLap }} · обновлено {{ updated }}</template>
        <template v-else-if="error">Нет связи с данными</template>
        <template v-else>Загрузка</template>
      </p>
      <BmChip v-if="error" variant="red">{{ error }}</BmChip>
    </BmCard>

    <div v-if="live" class="live__table-wrap">
      <table class="live__table">
        <thead>
          <tr class="display-sm">
            <th>P</th>
            <th>Пилот</th>
            <th class="live__hide-sm">Команда</th>
            <th class="live__num">Отрыв</th>
            <th class="live__num live__hide-sm">Интервал</th>
            <th class="live__num">Круг</th>
            <th class="live__num live__hide-sm">Лучший</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in live.drivers" :key="d.number">
            <td class="display-sm">{{ d.position || '—' }}</td>
            <td class="body-strong">
              {{ d.acronym }}
              <span class="body-sm live__hide-sm">{{ d.name }}</span>
            </td>
            <td class="body-sm live__hide-sm">{{ d.team }}</td>
            <td class="timing live__num">{{ formatGap(d.gap) }}</td>
            <td class="timing live__num live__hide-sm">{{ formatGap(d.interval) }}</td>
            <td class="timing live__num">{{ formatLap(d.lastLap) }}</td>
            <td class="timing live__num live__hide-sm">
              <BmChip v-if="d.number === bestLapNumber" variant="purple">{{ formatLap(d.bestLap) }}</BmChip>
              <template v-else>{{ formatLap(d.bestLap) }}</template>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>

<style scoped>
.live {
  display: grid;
  gap: var(--space-6);
}

.live__meta {
  margin-top: var(--space-2);
}

.live__table-wrap {
  background: var(--surface-sunken);
  border: var(--border-thick) solid var(--border);
  box-shadow: var(--shadow-hard);
  overflow-x: auto;
}

.live__table {
  width: 100%;
  border-collapse: collapse;
}

.live__table th,
.live__table td {
  padding: var(--space-3) var(--space-4);
  text-align: left;
  white-space: nowrap;
  vertical-align: middle;
}

.live__table thead th {
  color: var(--text-muted);
  border-bottom: var(--border-thin) solid var(--border);
}

.live__table tbody tr + tr td {
  border-top: 1px solid var(--border);
}

.live__num {
  text-align: right;
}

.live__hide-sm {
  display: none;
}

@media (min-width: 768px) {
  .live__hide-sm {
    display: table-cell;
  }

  span.live__hide-sm {
    display: inline;
    margin-left: var(--space-2);
  }
}
</style>
