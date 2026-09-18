<script setup lang="ts">
import { BmHeader } from '@/widgets/header'
import { useTheme } from '@/features/theme-toggle'
import { NextRace } from '@/widgets/next-race'
import { BmButton } from '@/shared/ui'
import { useRoute } from 'vue-router'
import { Moon, Sun } from 'lucide-vue-next'

const { theme, toggle } = useTheme()
const route = useRoute()
</script>

<template>
  <BmHeader>
    <NextRace />
    <RouterLink class="bm-btn bm-btn--ghost" to="/live">
      Live
    </RouterLink>
    <BmButton :aria-label="theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'" @click="toggle">
      <Sun v-if="theme === 'dark'" :size="20" :stroke-width="2.5" />
      <Moon v-else :size="20" :stroke-width="2.5" />
    </BmButton>
  </BmHeader>

  <main :class="route.meta.full ? 'bm-full' : 'bm-container'" style="padding-block: var(--space-8)">
    <RouterView />
  </main>
</template>

<style scoped>
.bm-full {
  padding-inline: var(--space-4);
}
</style>
