(function() {
  'use strict';
  const base = '/ulog/';
  const currentPath = window.location.pathname;
  const savedLang = localStorage.getItem('vitepress-lang');
  const supportedLangs = ['en', 'ru', 'zh'];
  if (savedLang && supportedLangs.includes(savedLang)) {
    const expectedPrefix = base + savedLang + '/';
    if (currentPath.startsWith(expectedPrefix)) {
      return;
    }
    let rest = '/';
    for (const l of supportedLangs) {
      const langPrefix = base + l + '/';
      if (currentPath.startsWith(langPrefix)) {
        rest = currentPath.substring(langPrefix.length - 1);
        break;
      }
    }
    if (currentPath === base || currentPath === base.slice(0, -1)) {
      rest = '/';
    }
    const newPath = base + savedLang + rest;
    window.location.replace(newPath);
    return;
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