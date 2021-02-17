export const expand = (string: string, envs: NodeJS.ProcessEnv) => {
  // eslint-disable-next-line no-template-curly-in-string
  const valuesRegex = new RegExp('\\${(.*?)}', 'g');
  return string.replace(valuesRegex, (matched: string) => {
    const varName = matched.substring(2, matched.length - 1);
    // expand to empty if varName not existent
    // alternatively, we can leave the ${} untouched - replace '' with matched.

    // varName format "ENVIRONMENT_NAME:defaultvalue"
    const [environmentName, ...rest] = varName.split(':');

    const envVariable = envs[environmentName];
    return envVariable !== undefined ? envVariable : rest.join(':');
  });
};

export const generateRoutes = (contents: string) => {
  return [];
};
