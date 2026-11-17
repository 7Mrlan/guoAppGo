// 主题取值与手动覆盖的读写。
export type Theme = "dark" | "light";

// 设置页的三个选项：跟随系统、固定亮色、固定暗色。
export type ThemePref = Theme | "system";

const THEME_KEY = "theme";
const prefersDark = matchMedia("(prefers-color-scheme: dark)");

// 读取手动覆盖，取值不合法时视为未覆盖。
export function storedPref(): ThemePref {
  const value = localStorage.getItem(THEME_KEY);
  return value === "dark" || value === "light" ? value : "system";
}

// 手动覆盖优先，其次系统偏好。
export function resolveTheme(pref: ThemePref): Theme {
  if (pref === "dark" || pref === "light") {
    return pref;
  }
  return prefersDark.matches ? "dark" : "light";
}

// 把生效主题写入 data-theme，CSS 令牌据此切换。
export function applyTheme(pref: ThemePref): Theme {
  const theme = resolveTheme(pref);
  document.documentElement.dataset.theme = theme;
  return theme;
}

// 写入覆盖；传 system 则清除覆盖、回到跟随系统。
export function setThemePref(pref: ThemePref): Theme {
  if (pref === "system") {
    localStorage.removeItem(THEME_KEY);
  } else {
    localStorage.setItem(THEME_KEY, pref);
  }
  return applyTheme(pref);
}

// 系统主题变化时，仅在未手动覆盖的情况下跟随。
export function watchSystemTheme(): void {
  prefersDark.addEventListener("change", () => {
    if (storedPref() === "system") {
      applyTheme("system");
    }
  });
}
