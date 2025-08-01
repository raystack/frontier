import React, { createContext, ReactNode, useContext } from 'react';
import { merge } from 'lodash';
import type { FrontierClientCustomizationOptions } from '../../shared/types';

const defaultCustomization: Required<FrontierClientCustomizationOptions> = {
  terminology: {
    organization: { singular: 'Organization', plural: 'Organizations' },
    project: { singular: 'Project', plural: 'Projects' },
    team: { singular: 'Team', plural: 'Teams' },
    member: { singular: 'Member', plural: 'Members' },
    user: { singular: 'User', plural: 'Users' },
    appName: 'Frontier'
  },
  messages: {
    billing: {},
    general: {}
  }
};

export const CustomizationContext =
  createContext<Required<FrontierClientCustomizationOptions>>(
    defaultCustomization
  );

export interface CustomizationProviderProps {
  children: ReactNode;
  config?: FrontierClientCustomizationOptions;
}

export const CustomizationProvider: React.FC<CustomizationProviderProps> = ({
  children,
  config = {}
}) => {
  const mergedConfig: Required<FrontierClientCustomizationOptions> = merge(
    {},
    defaultCustomization,
    config
  );

  return (
    <CustomizationContext.Provider value={mergedConfig}>
      {children}
    </CustomizationContext.Provider>
  );
};

const useCustomizationContext = () => {
  const context = useContext(CustomizationContext);
  return context || defaultCustomization;
};

export { defaultCustomization, useCustomizationContext };
