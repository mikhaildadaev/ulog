<template>
  <section class="redirect">
    <div class="container" :class="statusClass">
      <div class="wrapper">
        <svg class="loader" viewBox="0 0 50 50">
          <circle class="bg" cx="25" cy="25" r="20" />
          <circle class="progress-ring" cx="25" cy="25"  r="20" :style="{ strokeDashoffset: 125.6 - (125.6 * progress) / 100 }" />
        </svg>
      </div>
    </div>
  </section>
</template>

<script setup>
import { ref, onMounted } from 'vue'
const progress = ref(0)
const statusClass = ref('')
const redirectUrl = ref('')
const animateToProgress = (target) => {
  return new Promise((resolve) => {
    const start = progress.value
    const diff = target - start
    const duration = 200
    const startTime = Date.now()
    const step = () => {
      const elapsed = Date.now() - startTime
      const t = Math.min(elapsed / duration, 1)
      const eased = 1 - Math.pow(1 - t, 2)
      progress.value = Math.round(start + diff * eased)
      if (t < 1) {
        requestAnimationFrame(step)
      } else {
        progress.value = target
        resolve()
      }
    }
    step()
  })
}
const updateStatus = async (cls, progressTarget) => {
  statusClass.value = cls
  await animateToProgress(progressTarget)
}
const performRedirect = async () => {
  const base = import.meta.env.BASE_URL || ''
  const url = new URL(window.location.href)
  const path = window.location.pathname
  const langMatch = path.match(/^\/petly\/([a-z]{2})(?:\/|$)/)
  if (langMatch) {
    await updateStatus('status-success', 100)
    const savedLang = localStorage.getItem('vitepress-lang') || 'ru'
    window.location.href = `${base}${savedLang}/`
    return
  }
  const shortCode = url.searchParams.get('s')
  if (shortCode && shortCode.length > 5) {
    try {
      // ⚡ Этап 1: Начало загрузки (сразу 10%)
      await updateStatus('status-loading', 10)
      // ⚡ Этап 2: Загрузка index.json (10% → 40%)
      const indexResponse = await fetch('/petly/data/index.json')
      const index = await indexResponse.json()
      await updateStatus('status-loading', 40)
      // ⚡ Этап 3: Загрузка всех файлов (40% → 70%)
      const results = await Promise.allSettled(
        index.map(async ({ file, type, subtype }) => {
          const response = await fetch(`/petly/data/${file}`)
          if (!response.ok) throw new Error(`Failed to load ${file}`)
          const data = await response.json()
          return data.map(item => ({ ...item, _type: type, _subtype: subtype }))
        })
      )
      await updateStatus('status-loading', 70)
      // ⚡ Этап 4: Поиск питомца (70% → 90%)
      const allItems = results
        .filter(r => r.status === 'fulfilled')
        .flatMap(r => r.value)
      await updateStatus('status-loading', 90)
      // ⚡ Этап 5: Результат (90% → 100%)
      const item = allItems.find(p => p.short === shortCode)
      if (item) {
        const savedLang = localStorage.getItem('vitepress-lang') || 'ru'
        let redirectPath = `/${savedLang}/${item._type}/${item._subtype}/${item.uuid}`
        if (item.covenantID) {
          redirectPath = `/${savedLang}/${item._type}/${item.covenantID}/${item._subtype}/${item.uuid}`
        }
        redirectUrl.value = `${base}${redirectPath}`

        await updateStatus('status-success', 100)
        window.location.href = redirectUrl.value
      } else {
        await updateStatus('status-error', 100)
        const savedLang = localStorage.getItem('vitepress-lang') || 'ru'
        window.location.href = `${base}${savedLang}/`
      }
    } catch (error) {
      await updateStatus('status-error', 100)
      const savedLang = localStorage.getItem('vitepress-lang') || 'ru'
      window.location.href = `${base}${savedLang}/`
    }
  } else {
    await updateStatus('status-loading', 100)
    const savedLang = localStorage.getItem('vitepress-lang') || 'ru'
    window.location.href = `${base}${savedLang}/`
  }
}
onMounted(() => {
  performRedirect()
})
</script>