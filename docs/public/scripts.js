(function() {
  'use strict';
  const savedLang = localStorage.getItem('vitepress-lang');
  const currentPath = window.location.pathname;
  const supportedLangs = ['en', 'ru', 'zh'];
  const base = '/ulog/';
  function getRestPath(path) {
    for (const l of supportedLangs) {
      const prefix = base + l + '/';
      if (path.startsWith(prefix)) {
        return path.substring(prefix.length - 1);
      }
    }
    if (path.startsWith(base)) {
      return path.substring(base.length - 1);
    }
    return '/';
  }
  if (savedLang && supportedLangs.includes(savedLang)) {
    const targetPrefix = base + savedLang + '/';
    if (!currentPath.startsWith(targetPrefix)) {
      const rest = getRestPath(currentPath);
      window.location.replace(targetPrefix + rest.replace(/^\//, ''));
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