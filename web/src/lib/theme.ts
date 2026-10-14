export type Theme = 'light' | 'dark';

const key = 'theme';

export function currentTheme(): Theme {
  return document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light';
}

export function storedTheme(): Theme | null {
  try {
    const value = localStorage.getItem(key);
    return value === 'light' || value === 'dark' ? value : null;
  } catch {
    return null;
  }
}

export function applyTheme(theme: Theme, persist: boolean): void {
  document.documentElement.dataset.theme = theme;
  document.querySelector('meta[name="theme-color"]')?.setAttribute('content', theme === 'dark' ? '#142422' : '#ffffff');
  if (!persist) return;
  try {
    localStorage.setItem(key, theme);
  } catch {
    return;
  }
}
