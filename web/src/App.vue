<script setup lang="ts">
import { BmHeader } from '@/components/ui'
import { useTheme } from '@/composables/useTheme'
import { BmButton } from '@/components/ui'

const { toggle } = useTheme()
</script>

<template>
  <BmHeader>
    <BmButton variant="ghost" @click="toggle">
      Тема
    </BmButton>
    <RouterLink class="bm-btn bm-btn--ghost" to="/design-system">
      UI Kit
    </RouterLink>
  </BmHeader>

  <main class="bm-container" style="padding-block: var(--space-8)">
    <RouterView />
  </main>
</template>
