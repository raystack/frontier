import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { merge, keys, forEach } from 'lodash';
import en from './locales/en.json';
import type { TranslationResources } from './types';


const defaultResources: TranslationResources = {
  en: en,
};

const mergeResources = (resources?: TranslationResources): TranslationResources => {
  if (!resources) return defaultResources;
  
  return Object.keys(resources).reduce((acc, lang) => {
    acc[lang] = merge(
      {},
      defaultResources[lang] || {},
      resources[lang] || {}
    );
    return acc;
  }, {} as TranslationResources);
};

const initializeI18n = (mergedResources: TranslationResources, language: string) => {
  const i18nextResources = Object.keys(mergedResources).reduce((acc, lang) => {
    acc[lang] = { translation: mergedResources[lang] };
    return acc;
  }, {} as any);

  i18n
    .use(initReactI18next)
    .init({
      resources: i18nextResources,
      lng: language,
      fallbackLng: 'en',
      interpolation: {
        escapeValue: false,
      },
    });
};

const addResourceBundles = (mergedResources: TranslationResources) => {
  forEach(keys(mergedResources), (lang) => {
    i18n.addResourceBundle(lang, 'translation', mergedResources[lang], true, true);
  });
};

export const initI18n = (resources?: TranslationResources, language = 'en') => {
  const mergedResources = mergeResources(resources);
  
  if (!i18n.isInitialized) {
    initializeI18n(mergedResources, language);
  } else {
    addResourceBundles(mergedResources);
    i18n.changeLanguage(language);
  }
  
  return i18n;
};

initI18n();

export type { TranslationResources } from './types';
export default i18n;