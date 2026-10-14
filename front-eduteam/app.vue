<template>
  <div>
    <NuxtRouteAnnouncer />
    <NuxtPage />
  </div>
</template>

<style>
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap');
@import url('https://fonts.googleapis.com/css2?family=Red+Hat+Display:wght@400;500;600;700;800;900&display=swap');

:root {
  /* ── Corporate palette ── */
  --color-primary: #00A499;
  --color-primary-dark: #008a81;
  --color-primary-light: #e6f7f6;
  --color-accent: #EA7600;
  --color-accent-dark: #c96500;
  --color-accent-light: #fff3e6;
  --color-dark: #394049;
  --color-dark-soft: #4f5660;
  --color-text: #2b2f36;
  --color-text-secondary: #64707d;
  --color-border: #dfe4ea;
  --color-border-light: #eceff3;
  --color-surface: #ffffff;
  --color-background: #f4f6f9;
  --color-danger: #d7373f;
  --color-danger-light: #fceced;

  /* ── Typography ── */
  --font-family-main: "Inter", "Helvetica Neue", Arial, sans-serif;
  --font-family-titles: "Red Hat Display", "Inter", "Helvetica Neue", Arial, sans-serif;

  /* ── Shadows ── */
  --shadow-sm: 0 1px 2px rgba(57,64,73,0.06);
  --shadow-md: 0 2px 8px rgba(57,64,73,0.08);
  --shadow-lg: 0 4px 16px rgba(57,64,73,0.10);
  --shadow-xl: 0 8px 28px rgba(57,64,73,0.12);

  /* ── Radii ── */
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 14px;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: var(--font-family-main);
  line-height: 1.6;
  color: var(--color-text);
  background: var(--color-background);
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

h1, h2, h3, h4, h5, h6 {
  font-family: var(--font-family-titles);
  color: var(--color-dark);
}

input, select, button, textarea {
  font-family: inherit;
}

::selection {
  background: var(--color-primary);
  color: white;
}

/* ── Shared scrollbar ── */
::-webkit-scrollbar { width: 7px; }
::-webkit-scrollbar-track { background: var(--color-background); }
::-webkit-scrollbar-thumb { background: #c0c7d0; border-radius: 4px; }
::-webkit-scrollbar-thumb:hover { background: #a0a8b4; }
</style>
