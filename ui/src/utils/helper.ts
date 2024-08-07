import dayjs from "dayjs";
import type { TableColumnMetadata } from "~/types/types";
import { DEFAULT_DATE_FORMAT } from "./constants";

const DEFAULT_REDIRECT = "/";
const currencySymbolMap: Record<string, string> = {
  usd: "$",
  inr: "₹",
} as const;

const currencyDecimalMap: Record<string, number> = {
  usd: 2,
  inr: 2,
} as const;
const DEFAULT_DECIMAL = 0;

export const getCurrencyValue = (
  value: string = "",
  currency: string = "usd"
) => {
  const symbol =
    (currency?.toLowerCase() && currencySymbolMap[currency?.toLowerCase()]) ||
    currency;
  const currencyDecimal =
    (currency?.toLowerCase() && currencyDecimalMap[currency?.toLowerCase()]) ||
    DEFAULT_DECIMAL;

  const calculatedValue = (
    parseInt(value) / Math.pow(10, currencyDecimal)
  ).toFixed(currencyDecimal);
  const [intValue, decimalValue] = calculatedValue.toString().split(".");
  return [intValue, decimalValue, symbol];
};

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

export const keyToColumnMetaObject = (key: any) =>
  ({ key: key, name: key, value: key } as TableColumnMetadata);

/*
 * @desc returns date string - Eg, June 13, 2025. return '-' if the date in the argument is invalid.
 */
export const getFormattedDateString = (date: string) => date ? dayjs(date).format(DEFAULT_DATE_FORMAT) : '-'
