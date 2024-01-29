import { Flex } from '@raystack/apsara';

interface AmountProps {
  value: number;
  currency?: string;
  className?: string;
  currencyClassName?: string;
  decimalClassName?: string;
  valueClassName?: string;
}

const currenySymbolMap: Record<string, string> = {
  usd: '$',
  inr: '₹'
} as const;

const currenyDecimalMap: Record<string, number> = {
  usd: 2,
  inr: 2
} as const;

const DEFAULT_DECIMAL = 0;

export default function Amount({
  currency,
  value,
  className,
  currencyClassName,
  decimalClassName,
  valueClassName
}: AmountProps) {
  const symbol =
    (currency?.toLowerCase() && currenySymbolMap[currency?.toLowerCase()]) ||
    currency;
  const currencyDecimal =
    (currency?.toLowerCase() && currenyDecimalMap[currency?.toLowerCase()]) ||
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
        <span className={decimalClassName}>.{decimalValue}</span>
      </Flex>
    </Flex>
  );
}
