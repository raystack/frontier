import { Flex, Text } from "@odpf/apsara";
import { Cross1Icon } from "@radix-ui/react-icons";

type SheetHeaderProps = {
  title: string;
  onClick: () => void;
};

export function SheetHeader({ title, onClick }: SheetHeaderProps) {
  return (
    <Flex css={styles.header}>
      <Text size="4" css={{ fontWeight: "500" }}>
        {title}
      </Text>
      <Cross1Icon onClick={onClick} />
    </Flex>
  );
}

const styles = {
  header: {
    flexDirection: "row",
    justifyContent: "space-between",
    padding: "18px 32px",
    borderBottom: "1px solid $gray4",
  },
};
