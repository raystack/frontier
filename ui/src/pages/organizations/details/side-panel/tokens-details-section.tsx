import { List, Text, Flex } from "@raystack/apsara/v1";
import styles from "./side-panel.module.css";
import CoinIcon from "~/assets/icons/coin.svg?react";
import CoinColoredIcon from "~/assets/icons/coin-colored.svg?react";

export const TokensDetailsSection = ({
  organizationId,
}: {
  organizationId: string;
}) => {
  return (
    <List.Root>
      <List.Header>Tokens</List.Header>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Available tokens
        </List.Label>
        <List.Value>
          <Flex gap={3}>
            <CoinColoredIcon />
            <Text>-</Text>
          </Flex>
        </List.Value>
      </List.Item>
      <List.Item>
        <List.Label className={styles["side-panel-section-item-label"]}>
          Used till date
        </List.Label>
        <List.Value>
          <Flex gap={3}>
            <CoinIcon color="var(--rs-color-foreground-base-tertiary)" />
            <Text>-</Text>
          </Flex>
        </List.Value>
      </List.Item>
    </List.Root>
  );
};
