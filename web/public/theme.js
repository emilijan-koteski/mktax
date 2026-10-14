(() => {
  let stored = null;
  try {
    stored = localStorage.getItem('theme');
  } catch {
    stored = null;
  }
  const dark = stored === 'dark' || (stored !== 'light' && window.matchMedia('(prefers-color-scheme: dark)').matches);
  document.documentElement.dataset.theme = dark ? 'dark' : 'light';
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', dark ? '#142422' : '#ffffff');
})();
