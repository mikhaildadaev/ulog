(function() {
  'use strict';
  const savedLang = localStorage.getItem('vitepress-lang');
  const currentPath = window.location.pathname;
  const supportedLangs = ['en', 'ru', 'zh'];
  if (savedLang && supportedLangs.includes(savedLang)) {
    const prefix = '/' + savedLang + '/';
    if (!currentPath.startsWith(prefix)) {
      let rest = currentPath;
      for (const l of supportedLangs) {
        const p = '/' + l + '/';
        if (currentPath.startsWith(p)) {
          rest = currentPath.substring(p.length - 1);
          break;
        }
      }
      window.location.replace(prefix + rest.replace(/^\//, ''));
      return;
    }
  }
  const lang = document.documentElement.lang;
  if (lang) {
    localStorage.setItem('vitepress-lang', lang);
  }
  const observer = new MutationObserver(function() {
    const newLang = document.documentElement.lang;
    if (newLang) {
      localStorage.setItem('vitepress-lang', newLang);
    }
  });
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['lang']
  });
})();