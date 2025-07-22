export type TranslationResources = {
  [language: string]: {
    apiPlatform?: {
      appName?: string;
    };
    billing?: {
      plan_change?: Record<string, string>;
    };
    [key: string]: any;
  };
};