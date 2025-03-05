import { Cross1Icon } from "@radix-ui/react-icons";
import { Flex, Text } from "@raystack/apsara/v1";

type SheetHeaderProps = {
  title: string;
  onClick: () => void;
};

export function SheetHeader({ title, onClick }: SheetHeaderProps) {
  return (
    <Flex justify="between" style={styles.header}>
      <Text size={4} style={{ fontWeight: "500" }}>
        {title}
      </Text>
      <Cross1Icon
        onClick={onClick}
        style={{ cursor: "pointer" }}
        data-test-id="admin-ui-close-btn"
      />
    </Flex>
  );
}

const styles = {
  header: {
    padding: "18px 32px",
    borderBottom: "1px solid var(--border-base)",
  },
};
