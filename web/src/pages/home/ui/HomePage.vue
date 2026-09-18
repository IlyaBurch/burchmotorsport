<script setup lang="ts">
import { BmSectorBar, BmCard, BmButton } from '@/shared/ui'
</script>

<template>
  <section>
    <h1 class="display-xl">Lights Out</h1>
    <p class="body-lg" style="margin-top: var(--space-4)">
      Burch Motorsport — коммунити, телеметрия и трансляции Формулы 1.
    </p>
  </section>

  <BmSectorBar style="margin-block: var(--space-12)" />

  <section style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: var(--space-6)">
    <BmCard>
      <div class="bm-card__eyebrow">Live</div>
      <h2 class="bm-card__title">Трансляции</h2>
      <p class="body" style="margin-top: var(--space-3)">
        Смотри гонки с телеметрией на одном экране.
      </p>
      <BmButton style="margin-top: var(--space-4)">Скоро</BmButton>
    </BmCard>

    <BmCard>
      <div class="bm-card__eyebrow">Data</div>
      <h2 class="bm-card__title">Телеметрия</h2>
      <p class="body" style="margin-top: var(--space-3)">
        Позиции, скорости, отрывы — в реальном времени.
      </p>
      <BmButton style="margin-top: var(--space-4)">Скоро</BmButton>
    </BmCard>

    <BmCard>
      <div class="bm-card__eyebrow">Community</div>
      <h2 class="bm-card__title">Коммунити</h2>
      <p class="body" style="margin-top: var(--space-3)">
        Обсуждения, прогнозы, Fantasy F1.
      </p>
      <BmButton style="margin-top: var(--space-4)">Скоро</BmButton>
    </BmCard>
  </section>
</template>
