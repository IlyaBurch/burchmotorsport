<script setup lang="ts">
import { BmHeader } from '@/widgets/header'
import { useTheme } from '@/features/theme-toggle'
import { BmButton } from '@/shared/ui'
import { useRoute } from 'vue-router'

const { toggle } = useTheme()
const route = useRoute()
</script>

<template>
  <BmHeader>
    <BmButton variant="ghost" @click="toggle">
      Тема
    </BmButton>
    <RouterLink class="bm-btn bm-btn--ghost" to="/live">
      Live
    </RouterLink>
    <RouterLink class="bm-btn bm-btn--ghost" to="/design-system">
      UI Kit
    </RouterLink>
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
