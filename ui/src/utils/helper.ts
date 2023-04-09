const DEFAULT_REDIRECT = "/";

export function reduceByKey(data: Record<string, any>[], key: string) {
  return data.reduce((acc, value) => {
    return {
      ...acc,
      [value[key]]: value,
    };
  }, {});
}

export const capitalizeFirstLetter = (str: string) => {
  return str.charAt(0).toUpperCase() + str.slice(1);
};

// @ts-ignore
export const fetcher = (...args) => fetch(...args).then((res) => res.json());
