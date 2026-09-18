import { ref, watch, type Ref } from 'vue'
import { useRouter } from 'vue-router'
import { FIRST_SEASON, fetchMeetings, fetchSessions, type Live, type Meeting, type Session } from '@/shared/api/live'

/**
 * Season → grand prix → session selects. Only user actions navigate (session
 * goes into the url query); a loaded session just syncs the selects to itself.
 */
export function useSessionPicker(live: Ref<Live | null>) {
  const router = useRouter()

  const seasons = Array.from(
    { length: new Date().getFullYear() - FIRST_SEASON + 1 },
    (_, i) => FIRST_SEASON + i,
  ).reverse()

  const year = ref<number | null>(null)
  const meetingKey = ref<number | null>(null)
  const meetings = ref<Meeting[]>([])
  const sessions = ref<Session[]>([])

  const loadMeetings = async (y: number) => (meetings.value = await fetchMeetings(y))
  const loadSessions = async (mk: number) => (sessions.value = await fetchSessions(mk))
  const goTo = (key: number) =>
    router.replace({ query: { ...router.currentRoute.value.query, session: String(key) } })

  async function onYear(y: number) {
    year.value = y
    await loadMeetings(y)
    const last = meetings.value[meetings.value.length - 1]
    if (last) await onMeeting(last.meeting_key)
  }

  async function onMeeting(mk: number) {
    meetingKey.value = mk
    await loadSessions(mk)
    const last = sessions.value[sessions.value.length - 1]
    if (last) goTo(last.session_key)
  }

  watch(live, async (l) => {
    if (!l) return
    const y = new Date(l.session.date_start).getFullYear()
    if (y !== year.value) {
      year.value = y
      await loadMeetings(y)
    }
    if (l.session.meeting_key !== meetingKey.value) {
      meetingKey.value = l.session.meeting_key
      await loadSessions(meetingKey.value)
    }
  })

  return { seasons, year, meetingKey, meetings, sessions, onYear, onMeeting, goTo }
}
