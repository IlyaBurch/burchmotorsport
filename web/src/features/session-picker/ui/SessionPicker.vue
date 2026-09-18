<script setup lang="ts">
import { toRef } from 'vue'
import { sessionLabel, type Live } from '@/shared/api/live'
import { useSessionPicker } from '../model/useSessionPicker'

const props = defineProps<{ live: Live | null }>()
const { seasons, year, meetingKey, meetings, sessions, onYear, onMeeting, goTo } = useSessionPicker(toRef(props, 'live'))
</script>

<template>
  <!-- native <details>: folds away, works without js -->
  <details class="picker">
    <summary class="display-sm picker__summary">Другая сессия</summary>
    <div class="picker__grid">
      <label>
        <span class="display-sm">Сезон</span>
        <select class="picker__select body" :value="year ?? ''" @change="onYear(Number(($event.target as HTMLSelectElement).value))">
          <option v-for="y in seasons" :key="y" :value="y">{{ y }}</option>
        </select>
      </label>
      <label>
        <span class="display-sm">Гран-при</span>
        <select
          class="picker__select body"
          :disabled="!meetings.length"
          :value="meetingKey ?? ''"
          @change="onMeeting(Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="m in meetings" :key="m.meeting_key" :value="m.meeting_key">{{ m.meeting_name }}</option>
        </select>
      </label>
      <label>
        <span class="display-sm">Сессия</span>
        <select
          class="picker__select body"
          :disabled="!sessions.length"
          :value="live?.session.session_key ?? ''"
          @change="goTo(Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="s in sessions" :key="s.session_key" :value="s.session_key">{{ sessionLabel(s.session_name) }}</option>
        </select>
      </label>
    </div>
  </details>
</template>

<style scoped>
.picker {
  border-top: var(--border-thin) solid var(--border);
}

.picker__summary {
  display: flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 44px;
  cursor: pointer;
  color: var(--accent);
  list-style: none;
}

.picker__summary::-webkit-details-marker {
  display: none;
}

.picker__summary::before {
  content: '';
  width: 10px;
  height: 10px;
  border-right: 3px solid currentColor;
  border-bottom: 3px solid currentColor;
  transform: rotate(-45deg);
  transition: transform 100ms ease-out;
}

.picker[open] .picker__summary::before {
  transform: rotate(45deg);
}

.picker__grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
  gap: var(--space-3);
  padding-bottom: var(--space-2);
}

.picker__grid label {
  display: grid;
  gap: var(--space-2);
  min-width: 0;
}

/* input recipe from the guide, applied to a native select */
.picker__select {
  width: 100%;
  min-width: 0;
  min-height: 44px;
  padding: var(--space-3) var(--space-4);
  background: var(--surface-raised);
  color: var(--text);
  border: var(--border-thin) solid var(--border);
  border-radius: var(--radius-sm);
}

.picker__select:disabled {
  background: var(--surface-sunken);
  color: var(--text-muted);
}
</style>
