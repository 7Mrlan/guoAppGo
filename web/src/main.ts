import { createApp } from "vue";

import App from "./App.vue";
import "./style.css";
import { applyTheme, storedPref, watchSystemTheme } from "./theme";

applyTheme(storedPref());
watchSystemTheme();

createApp(App).mount("#app");
