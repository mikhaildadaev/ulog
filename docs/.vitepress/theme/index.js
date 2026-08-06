import DefaultTheme from 'vitepress/theme'
import './index.css'

import { watch, onMounted, nextTick } from 'vue'
import { useData, useRoute, useRouter } from 'vitepress'

import PageRedirect from '../components/PageRedirect.vue'

const base = import.meta.env.BASE_URL || ''

export default {
  extends: DefaultTheme,
  setup() {
    const { localeIndex } = useData()
    const route = useRoute()
    const router = useRouter()
    const supportedLangs = ['ru', 'en', 'zh']
    let isRedirecting = false
    let isShowingRedirect = false
    const isHomePage = (path) => {
      const clean = path.replace(/\/+$/, '')
      return clean === base || clean === base.slice(0, -1)
    }
    const cleanPath = (path) => {
      let cleaned = path
      cleaned = cleaned.replace(/\/+/g, '/')
      cleaned = cleaned.replace(/\/ly\//g, '/')
      cleaned = cleaned.replace(/^\/ly\//, '/')
      const parts = cleaned.split('/')
      let langFound = false
      const cleanParts = []
      for (const part of parts) {
        if (part && /^[a-z]{2}$/.test(part) && supportedLangs.includes(part)) {
          if (!langFound) {
            langFound = true
            cleanParts.push(part)
          }
        } else {
          cleanParts.push(part)
        }
      }
      let result = cleanParts.join('/')
      if (!result.startsWith(base)) {
        if (result.startsWith('/')) {
          result = base + result.slice(1)
        } else {
          result = base + result
        }
      }
      return result
    }
    const getLangFromPath = (path) => {
      let pathWithoutBase = path.replace(base, '')
      pathWithoutBase = pathWithoutBase.replace(/\.html$/, '')
      const segments = pathWithoutBase.split('/')
      for (const seg of segments) {
        if (seg && /^[a-z]{2}$/.test(seg) && supportedLangs.includes(seg)) {
          return seg
        }
      }
      return null
    }
    const getPathWithoutLang = (path) => {
      let pathWithoutBase = path.replace(base, '')
      const hasHtml = pathWithoutBase.endsWith('.html')
      let cleanPath = pathWithoutBase.replace(/\.html$/, '')
      const langPattern = new RegExp(`^(${supportedLangs.join('|')})/`)
      cleanPath = cleanPath.replace(langPattern, '')
      if (!cleanPath) {
        cleanPath = '/'
      } else if (!cleanPath.startsWith('/')) {
        cleanPath = '/' + cleanPath
      }
      return hasHtml ? cleanPath + '.html' : cleanPath
    }
    const ensureLanguageInStorage = () => {
      let storedLang = localStorage.getItem('vitepress-lang')
      if (!storedLang || !supportedLangs.includes(storedLang)) {
        storedLang = 'ru'
        localStorage.setItem('vitepress-lang', storedLang)
      }
      return storedLang
    }
    const syncUrlWithStorage = () => {
      if (isRedirecting) return
      const currentPath = route.path
      if (isHomePage(currentPath)) return false
      const cleanedPath = cleanPath(currentPath)
      if (cleanedPath !== currentPath) {
        isRedirecting = true
        router.go(cleanedPath)
        setTimeout(() => { isRedirecting = false }, 300)
        return true
      }
      const storedLang = localStorage.getItem('vitepress-lang')
      const urlLang = getLangFromPath(currentPath)
      if (!urlLang || urlLang !== storedLang) {
        const cleanPathPart = getPathWithoutLang(currentPath)
        const newPath = `${base}${storedLang}${cleanPathPart}`
        isRedirecting = true
        router.go(newPath)
        setTimeout(() => { isRedirecting = false }, 300)
        return true
      }
      return false
    }
    const updateLanguage = (newLang) => {
      if (!newLang || !supportedLangs.includes(newLang)) return
      if (isRedirecting) return
      const currentPath = route.path
      if (isHomePage(currentPath)) return
      const currentLang = getLangFromPath(currentPath)
      if (currentLang === newLang) return
      localStorage.setItem('vitepress-lang', newLang)
      const cleanPathPart = getPathWithoutLang(currentPath)
      const newPath = `${base}${newLang}${cleanPathPart}`
      isRedirecting = true
      router.go(newPath)
      setTimeout(() => {
        isRedirecting = false
      }, 300)
    }
    onMounted(() => {
      ensureLanguageInStorage()
      if (typeof window !== 'undefined') {
        const params = new URLSearchParams(window.location.search)
        if (params.has('s') && !isHomePage(route.path)) {
          isShowingRedirect = true
          setTimeout(() => {
            isShowingRedirect = false
            syncUrlWithStorage()
          }, 3000)
        } else {
          nextTick(() => {
            syncUrlWithStorage()
          })
        }
      }
    })
    watch(
      () => route.path,
      (newPath, oldPath) => {
        if (isRedirecting) return
        if (isHomePage(newPath)) return
        const cleanedPath = cleanPath(newPath)
        if (cleanedPath !== newPath) {
          isRedirecting = true
          router.go(cleanedPath)
          setTimeout(() => { isRedirecting = false }, 300)
          return
        }
        const storedLang = localStorage.getItem('vitepress-lang')
        const urlLang = getLangFromPath(newPath)
        if (!urlLang) {
          const cleanPathPart = getPathWithoutLang(newPath)
          const newFullPath = `${base}${storedLang}${cleanPathPart}`
          isRedirecting = true
          router.go(newFullPath)
          setTimeout(() => { isRedirecting = false }, 300)
          return
        }
        if (urlLang && urlLang !== storedLang) {
          const currentLocale = localeIndex.value
          if (currentLocale !== urlLang) {
            const cleanPathPart = getPathWithoutLang(newPath)
            const newFullPath = `${base}${storedLang}${cleanPathPart}`
            isRedirecting = true
            router.go(newFullPath)
            setTimeout(() => { isRedirecting = false }, 300)
          }
        }
      },
      { immediate: true }
    )
    watch(
      () => localeIndex.value,
      (newLocale, oldLocale) => {
        if (isRedirecting) return
        const currentPath = route.path
        if (isHomePage(currentPath)) return
        if (!newLocale || !supportedLangs.includes(newLocale)) return
        const storedLang = localStorage.getItem('vitepress-lang')
        const urlLang = getLangFromPath(currentPath)
        if (newLocale !== storedLang || newLocale !== urlLang) {
          localStorage.setItem('vitepress-lang', newLocale)
          if (newLocale !== urlLang) {
            const cleanPathPart = getPathWithoutLang(currentPath)
            const newPath = `${base}${newLocale}${cleanPathPart}`
            isRedirecting = true
            router.go(newPath)
            setTimeout(() => { isRedirecting = false }, 300)
          }
        }
      },
      { immediate: true }
    )
    return {}
  },
  enhanceApp({ app }) {
    app.component('PageRedirect', PageRedirect)
  },
}