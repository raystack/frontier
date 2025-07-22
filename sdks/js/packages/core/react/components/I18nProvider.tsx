import React, { useEffect } from 'react';
import { I18nextProvider } from 'react-i18next';
import { initI18n, TranslationResources } from '../i18n';

interface I18nProviderProps {
  children: React.ReactNode;
  resources?: TranslationResources;
  language?: string;
}

export const I18nProvider: React.FC<I18nProviderProps> = ({
  children,
  resources,
  language = 'en'
}) => {
  useEffect(() => {
    initI18n(resources, language);
  }, [resources, language]);

  const i18nInstance = initI18n(resources, language);

  return (
    <I18nextProvider i18n={i18nInstance}>
      {children}
    </I18nextProvider>
  );
};