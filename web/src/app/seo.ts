// One source of truth for per-page <title>/description.
// vite.config.ts bakes these into dist/<path>/index.html so Telegram and crawlers
// see them without JS; the router applies them on client-side navigation.
export const SEO: Record<string, { title: string; description: string }> = {
  '/': {
    title: 'Burch Motorsport',
    description: 'Коммунити Формулы 1: живая телеметрия, трансляции и обсуждения гонок.',
  },
  '/watch': {
    title: 'Трансляция с телеметрией · Burch Motorsport',
    description:
      'Смотри гонку Формулы 1 с RuTube и живой телеметрией на одном экране: карта трассы, race control, шины и зачёт рядом с видео или поверх него.',
  },
  '/live': {
    title: 'Живая телеметрия F1 · Burch Motorsport',
    description:
      'Тайминг Формулы 1 в реальном времени: позиции, отрывы, сектора, шины, карта трассы, race control и радио команд. Архив всех сессий с 2023 года.',
  },
}
