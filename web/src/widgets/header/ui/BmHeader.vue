<script setup lang="ts">
import { useTheme } from '@/features/theme-toggle'

const { theme } = useTheme()
defineEmits<{ logo: [] }>()
</script>

<template>
  <header class="bm-header">
    <RouterLink to="/" @click="$emit('logo')">
      <!-- Full lockup on desktop, simple mark on mobile -->
      <img
        v-if="theme !== 'light'"
        src="/brand/lockup-on-dark.svg"
        alt="Burch Motorsport"
        class="bm-header__logo bm-header__logo--full"
      />
      <img
        v-else
        src="/brand/lockup-on-light.svg"
        alt="Burch Motorsport"
        class="bm-header__logo bm-header__logo--full"
      />
      <img
        src="/brand/mark-simple.svg"
        alt="Burch Motorsport"
        class="bm-header__logo bm-header__logo--compact"
      />
    </RouterLink>

    <nav class="bm-header__nav">
      <slot />
    </nav>
  </header>
</template>

<style scoped>
.bm-header__logo--full {
  display: none;
}

.bm-header__logo--compact {
  display: block;
  height: 36px;
}

@media (min-width: 640px) {
  .bm-header__logo--full {
    display: block;
  }

  .bm-header__logo--compact {
    display: none;
  }
}

header a {
  text-decoration: none;
  display: flex;
  align-items: center;
}

header a:hover {
  background: none;
}
</style>
