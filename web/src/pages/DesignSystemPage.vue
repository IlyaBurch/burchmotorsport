<script setup lang="ts">
import { ref } from 'vue'
import {
  BmButton,
  BmCard,
  BmChip,
  BmInput,
  BmSectorBar,
  BmTabs,
  BmModal,
  BmToast,
} from '@/components/ui'

const activeTab = ref('race')
const tabs = [
  { key: 'race', label: 'Гонка' },
  { key: 'quali', label: 'Квала' },
  { key: 'practice', label: 'Практика' },
]

const inputValue = ref('')
const modalOpen = ref(false)
const toastVisible = ref(false)

function showToast() {
  toastVisible.value = true
  setTimeout(() => { toastVisible.value = false }, 2500)
}
</script>

<template>
  <h1 class="display-lg" style="margin-bottom: var(--space-8)">Design System</h1>

  <!-- Typography -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Типографика</h2>

    <div style="display: flex; flex-direction: column; gap: var(--space-4)">
      <p class="display-xl">Display XL — 72px</p>
      <p class="display-lg">Display LG — 48px</p>
      <p class="display-md">Display MD — 32px</p>
      <p class="display-sm">Display SM — 20px</p>
      <p class="body-lg">Body LG — 18px — Burch Motorsport — коммунити Формулы 1</p>
      <p class="body">Body — 16px — Основной текст интерфейса</p>
      <p class="body-strong">Body Strong — 16px/600 — Важный текст</p>
      <p class="body-sm">Body SM — 14px — Вторичный текст</p>
      <p class="timing-lg">1:19.327</p>
      <p class="timing">+0.184 | P3 | 58 кругов</p>
      <p class="timing-sm">S1 24.891 S2 28.443 S3 25.993</p>
    </div>
  </section>

  <BmSectorBar style="margin-bottom: var(--space-12)" />

  <!-- Colours -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Цвета</h2>

    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(120px, 1fr)); gap: var(--space-3)">
      <div style="background: var(--surface); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm">surface</span>
      </div>
      <div style="background: var(--surface-raised); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm">raised</span>
      </div>
      <div style="background: var(--surface-sunken); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm">sunken</span>
      </div>
      <div style="background: var(--accent); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm" style="color: var(--on-accent)">accent</span>
      </div>
      <div style="background: var(--blue-200); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm" style="color: var(--ink)">blue-200</span>
      </div>
      <div style="background: var(--blue-400); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm" style="color: var(--ink)">blue-400</span>
      </div>
      <div style="background: var(--blue-600); border: var(--border-thin); padding: var(--space-4)">
        <span class="body-sm" style="color: var(--paper)">blue-600</span>
      </div>
    </div>
  </section>

  <!-- Buttons -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Кнопки</h2>

    <div style="display: flex; flex-wrap: wrap; gap: var(--space-4); align-items: center">
      <BmButton variant="primary">Primary</BmButton>
      <BmButton>Default</BmButton>
      <BmButton variant="ghost">Ghost</BmButton>
      <BmButton disabled>Disabled</BmButton>
    </div>
  </section>

  <!-- Chips -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Плашки</h2>

    <div style="display: flex; flex-wrap: wrap; gap: var(--space-3)">
      <BmChip>Default</BmChip>
      <BmChip variant="purple">1:19.327</BmChip>
      <BmChip variant="green">PB 1:19.891</BmChip>
      <BmChip variant="yellow">+0.4s</BmChip>
      <BmChip variant="red">DNF</BmChip>
    </div>
  </section>

  <!-- Cards -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Карточки</h2>

    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: var(--space-6)">
      <BmCard>
        <div class="bm-card__eyebrow">Гран-при Монако</div>
        <h3 class="bm-card__title">Результаты</h3>
        <hr class="bm-card__divider" />
        <p class="bm-time">1:12.345</p>
        <p class="body-sm" style="margin-top: var(--space-1)">Лучший круг</p>
      </BmCard>

      <BmCard hoverable>
        <div class="bm-card__eyebrow">Live</div>
        <h3 class="bm-card__title">Hoverable</h3>
        <p class="body" style="margin-top: var(--space-3)">Наведи на карточку</p>
      </BmCard>
    </div>
  </section>

  <!-- Tabs -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Табы</h2>
    <BmTabs v-model="activeTab" :tabs="tabs" />
    <p class="body" style="margin-top: var(--space-4)">Выбрано: <strong>{{ activeTab }}</strong></p>
  </section>

  <!-- Input -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Поля ввода</h2>

    <div style="max-width: 400px; display: flex; flex-direction: column; gap: var(--space-4)">
      <BmInput v-model="inputValue" label="Позывной" placeholder="Введи имя" />
      <BmInput label="С ошибкой" placeholder="email" error="Неверный формат" />
    </div>
  </section>

  <!-- Table -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Таблица</h2>

    <div class="bm-table-wrap">
      <table class="bm-table">
        <thead>
          <tr>
            <th>Поз</th>
            <th>Пилот</th>
            <th>Команда</th>
            <th style="text-align: right">Время</th>
            <th style="text-align: right">Отрыв</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td class="display-sm">1</td>
            <td class="body-strong">VER</td>
            <td class="body">Red Bull</td>
            <td class="cell-number timing">1:19.327</td>
            <td class="cell-number timing">—</td>
          </tr>
          <tr class="is-current">
            <td class="display-sm">2</td>
            <td class="body-strong">NOR</td>
            <td class="body">McLaren</td>
            <td class="cell-number timing">1:19.511</td>
            <td class="cell-number timing">+0.184</td>
          </tr>
          <tr>
            <td class="display-sm">3</td>
            <td class="body-strong">LEC</td>
            <td class="body">Ferrari</td>
            <td class="cell-number timing">1:19.694</td>
            <td class="cell-number timing">+0.367</td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>

  <!-- Modal & Toast -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Модалка и тост</h2>

    <div style="display: flex; gap: var(--space-4)">
      <BmButton @click="modalOpen = true">Модалка</BmButton>
      <BmButton variant="ghost" @click="showToast">Тост</BmButton>
    </div>

    <BmModal :open="modalOpen" @close="modalOpen = false">
      <h3 class="display-md" style="margin-bottom: var(--space-4)">Box Box</h3>
      <p class="body">Пит-стоп подтверждён. Меняем на хард.</p>
      <BmButton variant="primary" style="margin-top: var(--space-6)" @click="modalOpen = false">
        Подтвердить
      </BmButton>
    </BmModal>

    <BmToast :visible="toastVisible">Состав сохранён</BmToast>
  </section>

  <!-- Skeleton -->
  <section style="margin-bottom: var(--space-12)">
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Скелетон</h2>

    <div style="display: flex; flex-direction: column; gap: var(--space-3); max-width: 400px">
      <div class="bm-skeleton" style="height: 24px; width: 60%"></div>
      <div class="bm-skeleton" style="height: 16px; width: 100%"></div>
      <div class="bm-skeleton" style="height: 16px; width: 80%"></div>
      <div class="bm-skeleton" style="height: 48px; width: 140px; margin-top: var(--space-2)"></div>
    </div>
  </section>

  <!-- Empty state -->
  <section>
    <h2 class="display-md" style="margin-bottom: var(--space-4)">Пустое состояние</h2>

    <div class="bm-empty" style="border: var(--border-thin)">
      <p class="bm-empty__title">Нет данных</p>
      <p class="bm-empty__text">Сессия ещё не началась</p>
      <BmButton variant="primary">Обновить</BmButton>
    </div>
  </section>
</template>
