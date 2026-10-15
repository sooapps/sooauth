import { createContext, useContext } from "react";
import { t, type Lang, type TranslateFn } from "./i18n";
import type { Data, Onboarding } from "./api";

export const I18nContext = createContext<{ lang: Lang; tr: TranslateFn; setLang: (lang: Lang) => void }>({
  lang: "en",
  tr: (key, vars) => t("en", key, vars),
  setLang: () => undefined,
});

export function useI18n() {
  return useContext(I18nContext);
}

export type AppState = {
  me: Data | null;
  projectId: string;
  notify: (message: string) => void;
  go: (section: string) => void;
  onboarding: Onboarding | null;
  refreshOnboarding: () => Promise<void>;
  openWizard: () => void;
};

export const AppContext = createContext<AppState>({
  me: null,
  projectId: "",
  notify: () => undefined,
  go: () => undefined,
  onboarding: null,
  refreshOnboarding: async () => undefined,
  openWizard: () => undefined,
});

export function useApp() {
  return useContext(AppContext);
}
