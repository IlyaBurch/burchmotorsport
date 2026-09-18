<script setup lang="ts">
import { BmHeader } from '@/widgets/header'
import { useTheme } from '@/features/theme-toggle'
import { BmButton } from '@/shared/ui'

const { toggle } = useTheme()
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

  <main class="bm-container" style="padding-block: var(--space-8)">
    <RouterView />
  </main>
</template>
