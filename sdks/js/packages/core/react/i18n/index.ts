import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import { merge, keys, forEach } from 'lodash';
import en from './locales/en.json';

export type TranslationResources = {
  [language: string]: {
    [namespace: string]: Record<string, any>;
  };
};

const defaultResources: TranslationResources = {
  en: {
    translation: en,
  },
};

const mergeResources = (resources?: TranslationResources): TranslationResources => {
  if (!resources) return defaultResources;
  
  return Object.keys(resources).reduce((acc, lang) => {
    acc[lang] = {
      translation: merge(
        {},
        defaultResources[lang]?.translation || {},
        resources[lang]?.translation || {}
      )
    };
    return acc;
  }, {} as TranslationResources);
};

const initializeI18n = (mergedResources: TranslationResources, language: string) => {
  i18n
    .use(initReactI18next)
    .init({
      resources: mergedResources,
      lng: language,
      fallbackLng: 'en',
      interpolation: {
        escapeValue: false,
      },
    });
};

const addResourceBundles = (mergedResources: TranslationResources) => {
  forEach(keys(mergedResources), (lang) => {
    forEach(keys(mergedResources[lang]), (namespace) => {
      i18n.addResourceBundle(lang, namespace, mergedResources[lang][namespace], true, true);
    });
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

export default i18n;