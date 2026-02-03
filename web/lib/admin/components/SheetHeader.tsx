import type { CSSProperties } from "react";
import { Cross1Icon } from "@radix-ui/react-icons";
import { Flex, Text } from "@raystack/apsara";

type SheetHeaderProps = {
  title: string;
  onClick: () => void;
  "data-testid"?: string;
};

export function SheetHeader({ title, onClick, "data-testid": dataTestId }: SheetHeaderProps) {
  return (
    <Flex justify="between" style={styles.header}>
      <Text size={4} style={{ fontWeight: "500" }}>
        {title}
      </Text>
      <Cross1Icon
        onClick={onClick}
        style={{ cursor: "pointer" }}
        data-testid={dataTestId ?? "admin-close-btn"}
      />
    </Flex>
  );
}

const styles: { header: CSSProperties } = {
  header: {
    padding: "18px 32px",
    borderBottom: "1px solid var(--rs-color-border-base-primary)",
  },
};
