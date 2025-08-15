import dayjs from "dayjs";
import type { TableColumnMetadata } from "~/types/types";
import { DEFAULT_DATE_FORMAT } from "./constants";
import { BillingAccountAddress } from "~/api/frontier";

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
  currency: string = "usd",
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
export const fetcher = (...args) => fetch(...args).then(res => res.json());

export const keyToColumnMetaObject = (key: any) =>
  ({ key: key, name: key, value: key }) as TableColumnMetadata;

/*
 * @desc returns date string - Eg, June 13, 2025. return '-' if the date in the argument is invalid.
 */
export const getFormattedDateString = (date: string) =>
  date ? dayjs(date).format(DEFAULT_DATE_FORMAT) : "-";

export const converBillingAddressToString = (
  address?: BillingAccountAddress,
) => {
  if (!address) return "";
  const { line1, line2, city, state, country, postal_code } = address;
  return [line1, line2, city, state, country, postal_code]
    .filter(v => v)
    .join(", ");
};

/**
 * @desc Download a file
 * @param data - The file to download
 * @param filename - The name of the file to download
 */
export const downloadFile = (data: File | Blob, filename: string) => {
  const link = document.createElement("a");
  const downloadUrl = window.URL.createObjectURL(
    data instanceof Blob ? data : new Blob([data]),
  );
  link.href = downloadUrl;
  link.setAttribute("download", filename);
  document.body.appendChild(link);
  link.click();
  link.parentNode?.removeChild(link);
  window.URL.revokeObjectURL(downloadUrl);
};

/**
 * @desc Export CSV data from a Connect streaming service method
 * @param streamingMethod - The streaming method from the client
 * @param request - The request object
 * @param filename - The name of the CSV file to download
 */
export const exportCsvFromStream = async <T>(
  streamingMethod: (request: T) => AsyncIterable<{ data?: Uint8Array }>,
  request: T,
  filename: string,
) => {
  const chunks: Uint8Array[] = [];

  for await (const response of streamingMethod(request)) {
    if (response.data) {
      chunks.push(response.data);
    }
  }

  const blob = new Blob(chunks, { type: "text/csv" });
  downloadFile(blob, filename);
};
