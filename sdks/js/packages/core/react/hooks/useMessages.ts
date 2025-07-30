import { useContext } from 'react';
import { get } from 'lodash';
import { CustomizationContext, defaultCustomization } from '../contexts/CustomizationContext';

export const useMessages = () => {
  const context = useContext(CustomizationContext);
  const config = context || defaultCustomization;
  const { messages } = config;

  const getMessage = (path: string): string | undefined => {
    const value = get(messages, path);
    return typeof value === 'string' ? value : undefined;
  };

  return { getMessage };
};