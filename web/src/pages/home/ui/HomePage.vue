<script setup lang="ts">
import { computed, ref } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { BmSectorBar, BmCard, BmChip } from '@/shared/ui'
import { DriverPlate } from '@/entities/session'
import {
  countdown,
  fetchLive,
  fetchNext,
  fetchSessions,
  formatGap,
  formatLap,
  formatMsk,
  sessionLabel,
  type Live,
  type NextSession,
  type Session,
} from '@/shared/api/live'

// ponytail: page-level fetches, all served from the api cache. Move into an
// entity when a second page needs the weekend or the last race.
const next = ref<NextSession | null>(null)
const weekend = ref<Session[]>([])
const last = ref<Live | null>(null)

fetchNext()
  .then(async (n) => {
    next.value = n
    if (n) weekend.value = await fetchSessions((await fetchLive(n.session_key)).session.meeting_key)
  })
  .catch(() => {})
fetchLive('latest')
  .then((l) => (last.value = l.upcoming ? null : l))
  .catch(() => {})

const now = ref(Date.now())
useIntervalFn(() => (now.value = Date.now()), 30_000)

const podium = computed(() => last.value?.drivers.slice(0, 3) ?? [])
const bestLap = computed(() => {
  const d = last.value?.drivers.filter((x) => x.bestLap != null) ?? []
  return d.length ? d.reduce((a, b) => (a.bestLap! <= b.bestLap! ? a : b)) : null
})
</script>

<template>
  <section class="home__hero">
    <!-- brand plate: flat blue-600, mark as an image, paper text on top (guide §2) -->
    <div class="home__plate">
      <img src="/brand/mark.svg" alt="Burch Motorsport" class="home__mark" />
      <div class="home__plate-text">
        <div class="display-sm home__kicker">Burch Motorsport</div>
        <h1 class="display-xl">Личный пит-уолл</h1>
        <p class="body-lg home__lead">Телеметрия, трансляции и коммунити Формулы 1. Всё, что видит команда на пит-уолле, у тебя на экране.</p>
        <div class="home__actions">
          <RouterLink class="bm-btn bm-btn--primary" to="/watch">Смотреть трансляцию</RouterLink>
          <RouterLink class="bm-btn home__btn-paper" to="/live">Телеметрия</RouterLink>
        </div>
      </div>
    </div>

    <BmCard v-if="next" class="home__next">
      <div class="bm-card__eyebrow">Ближайшая сессия</div>
      <p class="display-lg home__countdown">{{ countdown(Date.parse(next.date_start) - now) }}</p>
      <p class="body-strong">{{ sessionLabel(next.session_name) }} · {{ next.location }}, {{ next.country_name }}</p>
      <p class="body-sm">{{ formatMsk(next.date_start) }}</p>
      <ol v-if="weekend.length" class="home__weekend">
        <li v-for="s in weekend" :key="s.session_key" :class="{ 'home__weekend--past': Date.parse(s.date_end) < now }">
          <RouterLink class="body-strong home__weekend-link" :to="{ path: '/live', query: { session: String(s.session_key) } }">
            {{ sessionLabel(s.session_name) }}
          </RouterLink>
          <span class="timing-sm">{{ formatMsk(s.date_start) }}</span>
        </li>
      </ol>
    </BmCard>
  </section>

  <BmSectorBar class="home__bar" />

  <section class="home__grid">
    <BmCard class="home__feature">
      <div class="bm-card__eyebrow">Live</div>
      <h2 class="bm-card__title">Трансляции</h2>
      <p class="body">Видео с RuTube и телеметрия на одном экране: карта, race control, шины, зачёт.</p>
      <RouterLink class="bm-btn home__cta" to="/watch">Смотреть</RouterLink>
    </BmCard>

    <BmCard class="home__feature">
      <div class="bm-card__eyebrow">Data</div>
      <h2 class="bm-card__title">Телеметрия</h2>
      <p class="body">Позиции, отрывы, сектора, стратегия шин, радио команд. Архив всех сессий с 2023 года.</p>
      <RouterLink class="bm-btn home__cta" to="/live">Открыть</RouterLink>
    </BmCard>

    <BmCard class="home__feature">
      <div class="bm-card__eyebrow">Community</div>
      <h2 class="bm-card__title">Коммунити</h2>
      <p class="body">Обсуждения, прогнозы, Fantasy F1.</p>
      <span class="bm-btn home__cta" aria-disabled="true">Скоро</span>
    </BmCard>

    <BmCard v-if="last" class="home__last">
      <div class="bm-card__eyebrow">Последняя сессия</div>
      <h2 class="bm-card__title">{{ last.session.circuit_short_name }} · {{ sessionLabel(last.session.session_name) }}</h2>
      <ol class="home__podium">
        <li v-for="d in podium" :key="d.number">
          <span class="display-sm">{{ d.position }}</span>
          <DriverPlate :label="d.acronym" :colour="d.teamColour" />
          <span class="body-sm home__name">{{ d.name }}</span>
          <span class="timing">{{ formatGap(d.gap) }}</span>
        </li>
      </ol>
      <p v-if="bestLap" class="home__best">
        <span class="body-sm">Лучший круг</span>
        <DriverPlate :label="bestLap.acronym" :colour="bestLap.teamColour" />
        <BmChip variant="purple">{{ formatLap(bestLap.bestLap) }}</BmChip>
      </p>
      <RouterLink class="bm-btn home__cta" :to="{ path: '/live', query: { session: String(last.session.session_key) } }">
        Разбор сессии
      </RouterLink>
    </BmCard>
  </section>
</template>

<style scoped>
.home__hero {
  display: grid;
  gap: var(--space-6);
  align-items: start;
}

@media (min-width: 1024px) {
  .home__hero {
    grid-template-columns: minmax(0, 1fr) 380px;
    align-items: center;
  }
}

.home__plate {
  display: grid;
  gap: var(--space-4);
  padding: var(--space-6);
  background: var(--blue-600);
  color: var(--paper);
  border: var(--border-thick) solid var(--border);
  box-shadow: var(--shadow-hard-lg);
}

@media (min-width: 768px) {
  .home__plate {
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: var(--space-8);
    padding: var(--space-8);
    min-height: 100%;
    box-sizing: border-box;
  }
}

.home__mark {
  width: 96px;
  height: auto;
}

@media (min-width: 768px) {
  .home__mark {
    width: 180px;
  }
}

.home__kicker {
  color: var(--blue-200);
}

.home__lead {
  margin-top: var(--space-3);
  max-width: 44ch;
}

/* secondary button on the blue plate: paper fill so it reads against blue-600 */
.home__btn-paper {
  background: var(--paper);
  color: var(--ink);
}

.home__actions {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  margin-top: var(--space-6);
}

.home__countdown {
  margin: var(--space-2) 0 var(--space-3);
}

.home__weekend {
  list-style: none;
  margin: var(--space-4) 0 0;
  padding: var(--space-3) 0 0;
  border-top: var(--border-thin) solid var(--border);
  display: grid;
  gap: var(--space-2);
}

.home__weekend li {
  max-width: none;
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: var(--space-3);
}

.home__weekend--past {
  color: var(--text-muted);
}

.home__weekend-link {
  color: inherit;
  text-decoration: none;
}

.home__weekend-link:hover {
  color: var(--on-accent);
}

.home__bar {
  margin-block: var(--space-12);
}

.home__grid {
  display: grid;
  gap: var(--space-6);
  grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
}

/* content on top, cta pinned to the bottom edge of every card */
.home__feature,
.home__last {
  display: flex;
  flex-direction: column;
  gap: var(--space-3);
}

.home__cta {
  margin-top: auto;
  align-self: flex-start;
}

.home__last {
  grid-column: 1 / -1;
}

@media (min-width: 1024px) {
  .home__grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

.home__podium {
  list-style: none;
  margin: 0;
  padding: 0;
  display: grid;
  gap: var(--space-2);
}

.home__podium li {
  max-width: none;
  display: grid;
  grid-template-columns: 24px 56px 1fr auto;
  gap: var(--space-3);
  align-items: center;
}

.home__best {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin: 0;
  padding-top: var(--space-3);
  border-top: var(--border-thin) solid var(--border);
}

@media (min-width: 1024px) {
  .home__podium {
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: var(--space-4);
  }

  .home__podium li {
    padding: var(--space-3);
    background: var(--surface-sunken);
    border: var(--border-thin) solid var(--border);
  }
}
</style>
