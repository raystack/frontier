import type React from 'react';

export type AdminConfig = {
  title?: string;
  logo?: string;
};

export type AdminLoginProps = {
  config?: AdminConfig;
  logoIcon?: React.ReactNode;
};

