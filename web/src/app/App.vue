<script setup lang="ts">
import { BmHeader } from '@/widgets/header'
import { useTheme } from '@/features/theme-toggle'
import { NextRace } from '@/widgets/next-race'
import { useNightLords } from '@/features/night-lords'
import { BmButton, BmToast } from '@/shared/ui'
import { useRoute } from 'vue-router'
import { Moon, Sun } from 'lucide-vue-next'

const { theme, toggle } = useTheme()
const route = useRoute()
const egg = useNightLords()
</script>

<template>
  <BmHeader @logo="egg.tap()">
    <NextRace />
    <RouterLink class="bm-btn bm-btn--ghost" to="/watch">
      Трансляция
    </RouterLink>
    <RouterLink class="bm-btn bm-btn--ghost" to="/live">
      Телеметрия
    </RouterLink>
    <BmButton :aria-label="theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'" @click="toggle">
      <Sun v-if="theme === 'dark'" :size="20" :stroke-width="2.5" />
      <Moon v-else :size="20" :stroke-width="2.5" />
    </BmButton>
  </BmHeader>

  <main :class="route.meta.full ? 'bm-full' : 'bm-container'" style="padding-block: var(--space-8)">
    <RouterView />
  </main>

  <BmToast :visible="egg.toast.value">
    <span class="display-sm">{{ egg.active.value ? 'Ave Dominus Nox' : 'Возвращаемся в свет' }}</span>
  </BmToast>
</template>

<style scoped>
.bm-full {
  padding-inline: var(--space-4);
}
</style>
