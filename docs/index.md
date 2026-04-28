---
layout: false
---

<script setup>
import { onMounted } from 'vue'

onMounted(() => {
  const savedLang = localStorage.getItem('vitepress-lang')
  const supportedLangs = ['en', 'ru', 'zh']
  const lang = supportedLangs.includes(savedLang) ? savedLang : 'en'
  window.location.replace(`/${lang}/`)
})
</script>