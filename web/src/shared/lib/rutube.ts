/** Pull the video id out of any RuTube url or a bare id. null if it is not one. */
export function parseRutubeId(input: string): string | null {
  const s = input.trim()
  if (/^[0-9a-f]{32}$/i.test(s)) return s.toLowerCase()
  try {
    const u = new URL(s)
    if (!/(^|\.)rutube\.ru$/.test(u.hostname)) return null
    const m = /\/(?:video|play\/embed|live\/video)\/([0-9a-f]{32})/i.exec(u.pathname)
    return m ? m[1]!.toLowerCase() : null
  } catch {
    return null
  }
}

// getPlayOptions=title makes the player post the video title (player:playOptionLoaded)
export const rutubeEmbedUrl = (id: string) =>
  `https://rutube.ru/play/embed/${id}?autostartmute=true&autoplay=1&getPlayOptions=title`
