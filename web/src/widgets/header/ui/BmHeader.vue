<script setup lang="ts">
import { useTheme } from '@/features/theme-toggle'

const { theme } = useTheme()
defineEmits<{ logo: [] }>()
</script>

<template>
  <header class="bm-header">
    <!-- Night Lords: the winged skull and the greeting stand in for the lockup -->
    <RouterLink v-if="theme === 'night-lords'" to="/" class="bm-btn bm-header__nl" @click="$emit('logo')">
      <img src="/brand/night-lords.png" alt="Night Lords" class="bm-header__nl-emblem" />
      <span class="bm-header__nl-text">Ave Dominus Nox</span>
    </RouterLink>
    <RouterLink v-else to="/" @click="$emit('logo')">
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

.bm-header__nl {
  gap: var(--space-3);
  padding-block: var(--space-1);
}

.bm-header__nl-emblem {
  height: 36px;
  width: auto;
}

.bm-header__nl-text {
  display: none;
}

@media (min-width: 640px) {
  .bm-header__nl-text {
    display: inline;
  }
}

header a:hover {
  background: none;
}
</style>
