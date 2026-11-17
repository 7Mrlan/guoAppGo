import { createApp } from "vue";

import App from "./App.vue";
import "./style.css";

// 生效中的主题取值。
type Theme = "dark" | "light";

const THEME_KEY = "theme";
const prefersDark = matchMedia("(prefers-color-scheme: dark)");

// 手动覆盖优先，取值不合法时回落到系统偏好。
function storedTheme(): Theme | null {
  const value = localStorage.getItem(THEME_KEY);
  return value === "dark" || value === "light" ? value : null;
}

// 按手动覆盖优先、其次系统偏好写入 data-theme。
function applyTheme(): void {
  const theme: Theme = storedTheme() ?? (prefersDark.matches ? "dark" : "light");
  document.documentElement.dataset.theme = theme;
}

applyTheme();
prefersDark.addEventListener("change", () => {
  if (storedTheme() === null) {
    applyTheme();
  }
});

createApp(App).mount("#app");
