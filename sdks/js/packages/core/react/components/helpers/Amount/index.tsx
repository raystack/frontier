import { Flex } from '@raystack/apsara/v1';

interface AmountProps {
  value: number;
  currency?: string;
  className?: string;
  currencyClassName?: string;
  decimalClassName?: string;
  valueClassName?: string;
  hideDecimals?: boolean;
}

const currencySymbolMap: Record<string, string> = {
  usd: '$',
  inr: '₹'
} as const;

const currencyDecimalMap: Record<string, number> = {
  usd: 2,
  inr: 2
} as const;

const DEFAULT_DECIMAL = 0;

/*
 * @deprecated Use Amount component from @raystack/apsara/v1 instead.
 */
export default function Amount({
  currency = 'usd',
  value = 0,
  className,
  currencyClassName,
  decimalClassName,
  valueClassName,
  hideDecimals
}: AmountProps) {
  const symbol =
    (currency?.toLowerCase() && currencySymbolMap[currency?.toLowerCase()]) ||
    currency;
  const currencyDecimal =
    (currency?.toLowerCase() && currencyDecimalMap[currency?.toLowerCase()]) ||
    DEFAULT_DECIMAL;

  const calculatedValue = (value / Math.pow(10, currencyDecimal)).toFixed(
    currencyDecimal
  );
  const [intValue, decimalValue] = calculatedValue.toString().split('.');

  return (
    <Flex className={className}>
      <span className={currencyClassName}>{symbol}</span>
      <Flex className={valueClassName}>
        <span>{intValue}</span>
        {hideDecimals ? null : (
          <span className={decimalClassName}>.{decimalValue}</span>
        )}
      </Flex>
    </Flex>
  );
}
