<script setup lang="ts">
import { computed, ref } from 'vue'
import { useIntervalFn } from '@vueuse/core'
import { BmSectorBar, BmCard, BmChip } from '@/shared/ui'
import { DriverPlate } from '@/entities/session'
import { useTheme } from '@/features/theme-toggle'
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

const { theme } = useTheme()
const nl = computed(() => theme.value === 'night-lords')

// ponytail: two copies of the home page copy, picked by theme. If a third
// theme ever wants its own words, move this into shared/i18n.
const copy = computed(() =>
  nl.value
    ? {
        kicker: 'VIII легион · Нострамо',
        title: 'Пит-уолл Ночного Охотника',
        lead: 'Ауспекс, вокс и пикт-каналы Астартес. Всё, что видит Легион с орбиты, у тебя на когитаторе.',
        watch: 'Открыть пикт-канал',
        live: 'Ауспекс',
        nextTitle: 'Следующая охота',
        f1: { eyebrow: 'Вокс', title: 'Пикт-трансляции', text: 'Картинка с RuTube и ауспекс на одном экране: карта сектора, приказы командования, шины, зачёт.', cta: 'Смотреть' },
        f2: { eyebrow: 'Ауспекс', title: 'Телеметрия', text: 'Позиции, отрывы, сектора, стратегия шин, вокс-перехваты команд. Архив всех охот с 2023 года.', cta: 'Открыть' },
        f3: { eyebrow: 'Легион', title: 'Братство', text: 'Обсуждения, пророчества, Fantasy F1. Страх сильнее веры.', cta: 'Скоро' },
        lastTitle: 'Последняя охота',
        bestLap: 'Быстрейший из братьев',
        review: 'Хроники охоты',
      }
    : {
        kicker: 'Burch Motorsport',
        title: 'Личный пит-уолл',
        lead: 'Телеметрия, трансляции и коммунити Формулы 1. Всё, что видит команда на пит-уолле, у тебя на экране.',
        watch: 'Смотреть трансляцию',
        live: 'Телеметрия',
        nextTitle: 'Ближайшая сессия',
        f1: { eyebrow: 'Live', title: 'Трансляции', text: 'Видео с RuTube и телеметрия на одном экране: карта, race control, шины, зачёт.', cta: 'Смотреть' },
        f2: { eyebrow: 'Data', title: 'Телеметрия', text: 'Позиции, отрывы, сектора, стратегия шин, радио команд. Архив всех сессий с 2023 года.', cta: 'Открыть' },
        f3: { eyebrow: 'Community', title: 'Коммунити', text: 'Обсуждения, прогнозы, Fantasy F1.', cta: 'Скоро' },
        lastTitle: 'Последняя сессия',
        bestLap: 'Лучший круг',
        review: 'Разбор сессии',
      },
)
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
    <div class="home__plate bm-card">
      <img src="/brand/mark.svg" alt="Burch Motorsport" class="home__mark" />
      <div class="home__plate-text">
        <div class="display-sm home__kicker">{{ copy.kicker }}</div>
        <h1 class="display-xl">{{ copy.title }}</h1>
        <p class="body-lg home__lead">{{ copy.lead }}</p>
        <div class="home__actions">
          <RouterLink class="bm-btn bm-btn--primary" to="/watch">{{ copy.watch }}</RouterLink>
          <RouterLink class="bm-btn" to="/live">{{ copy.live }}</RouterLink>
        </div>
      </div>
    </div>

    <BmCard v-if="next" class="home__next">
      <div class="bm-card__eyebrow">{{ copy.nextTitle }}</div>
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

  <section v-if="nl" class="home__legion bm-card">
    <img src="/brand/night-lords.png" alt="Night Lords" class="home__legion-emblem" />
    <div>
      <div class="bm-card__eyebrow">VIII легион</div>
      <h2 class="display-lg">Ave Dominus Nox</h2>
      <p class="display-md">Нас ведёт Конрад Керз</p>
      <p class="body-lg">Мы не просим верности. Мы приходим ночью, и страх делает остальное.</p>
      <p class="body-sm">Повелители Ночи. Нострамо помнит.</p>
    </div>
  </section>

  <BmSectorBar class="home__bar" />

  <section class="home__grid">
    <BmCard class="home__feature">
      <div class="bm-card__eyebrow">{{ copy.f1.eyebrow }}</div>
      <h2 class="bm-card__title">{{ copy.f1.title }}</h2>
      <p class="body">{{ copy.f1.text }}</p>
      <RouterLink class="bm-btn home__cta" to="/watch">{{ copy.f1.cta }}</RouterLink>
    </BmCard>

    <BmCard class="home__feature">
      <div class="bm-card__eyebrow">{{ copy.f2.eyebrow }}</div>
      <h2 class="bm-card__title">{{ copy.f2.title }}</h2>
      <p class="body">{{ copy.f2.text }}</p>
      <RouterLink class="bm-btn home__cta" to="/live">{{ copy.f2.cta }}</RouterLink>
    </BmCard>

    <BmCard class="home__feature">
      <div class="bm-card__eyebrow">{{ copy.f3.eyebrow }}</div>
      <h2 class="bm-card__title">{{ copy.f3.title }}</h2>
      <p class="body">{{ copy.f3.text }}</p>
      <span class="bm-btn home__cta" aria-disabled="true">{{ copy.f3.cta }}</span>
    </BmCard>

    <BmCard v-if="last" class="home__last">
      <div class="bm-card__eyebrow">{{ copy.lastTitle }}</div>
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
        <span class="body-sm">{{ copy.bestLap }}</span>
        <DriverPlate :label="bestLap.acronym" :colour="bestLap.teamColour" />
        <BmChip variant="purple">{{ formatLap(bestLap.bestLap) }}</BmChip>
      </p>
      <RouterLink class="bm-btn home__cta" :to="{ path: '/live', query: { session: String(last.session.session_key) } }">
        {{ copy.review }}
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
    align-items: stretch; /* both cards share the row height */
  }
}

.home__plate {
  display: grid;
  gap: var(--space-4);
}

@media (min-width: 768px) {
  .home__plate {
    grid-template-columns: auto minmax(0, 1fr);
    align-items: center;
    gap: var(--space-8);
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
  color: var(--accent);
}

.home__lead {
  margin-top: var(--space-3);
  max-width: 44ch;
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

.home__legion {
  display: grid;
  gap: var(--space-6);
  margin-top: var(--space-6);
  align-items: center;
}

@media (min-width: 768px) {
  .home__legion {
    grid-template-columns: 240px minmax(0, 1fr);
  }
}

.home__legion-emblem {
  width: 100%;
  max-width: 240px;
  height: auto;
  justify-self: center;
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
