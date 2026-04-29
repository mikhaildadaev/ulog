(function() {
  'use strict';
  const base = '/ulog/';
  const currentPath = window.location.pathname;
  const savedLang = localStorage.getItem('vitepress-lang');
  const supportedLangs = ['en', 'ru', 'zh'];
  const targetLang = (savedLang && supportedLangs.includes(savedLang)) 
    ? savedLang 
    : 'en';
  let needsRedirect = false;
  let rest = '/';
  const expectedPrefix = base + targetLang + '/';
  if (!currentPath.startsWith(expectedPrefix)) {
    needsRedirect = true;
  }
  if (currentPath === base || currentPath === base.slice(0, -1)) {
    needsRedirect = true;
  }
  if (needsRedirect) {
    for (const l of supportedLangs) {
      const langPrefix = base + l + '/';
      if (currentPath.startsWith(langPrefix)) {
        rest = currentPath.substring(langPrefix.length - 1);
        break;
      }
    }
    const newPath = base + targetLang + rest;
    window.location.replace(newPath);
    return;
  }
  const lang = document.documentElement.lang;
  if (lang && supportedLangs.includes(lang)) {
    localStorage.setItem('vitepress-lang', lang);
  }
  const observer = new MutationObserver(function() {
    const newLang = document.documentElement.lang;
    if (newLang && supportedLangs.includes(newLang)) {
      localStorage.setItem('vitepress-lang', newLang);
    }
  });
  observer.observe(document.documentElement, {
    attributes: true,
    attributeFilter: ['lang']
  });
})();