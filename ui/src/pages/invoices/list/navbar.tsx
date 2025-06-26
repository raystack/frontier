import { Flex, Text } from "@raystack/apsara/v1";
import styles from "./list.module.css";
import InvoicesIcon from "~/assets/icons/invoices.svg?react";

export const InvoicesNavabar = () => {
  return (
    <nav className={styles.navbar}>
      <Flex gap={2}>
        <InvoicesIcon />
        <Text size={2} weight={500}>
          Invoices
        </Text>
      </Flex>
    </nav>
  );
};
