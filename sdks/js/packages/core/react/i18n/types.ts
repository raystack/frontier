export type TranslationResources = {
  [language: string]: {
    apiPlatform?: {
      appName?: string;
    };
    billing?: {
      plan_change?: Record<string, string>;
    };
    terminology?: {
      organization?: string;
      [key: string]: any;
    };
    [key: string]: any;
  };
};