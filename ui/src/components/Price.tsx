import { Flex } from "@raystack/apsara/v1";
import { getCurrencyValue } from "~/utils/helper";

type PriceProps = {
  value?: string;
  currency?: string;
};
export const Price = ({ value = "", currency = "usd" }: PriceProps) => {
  const [intValue, decimalValue, symbol] = getCurrencyValue(value, currency);
  return (
    <Flex>
      <span>{symbol}</span>
      <Flex>
        <span>{intValue}</span>
        <span>.{decimalValue}</span>
      </Flex>
    </Flex>
  );
};
