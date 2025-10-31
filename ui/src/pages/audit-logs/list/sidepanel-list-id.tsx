import { CopyButton, Flex, List, Text, Tooltip } from "@raystack/apsara";
import styles from "./list.module.css";

export default function SidepanelListId({ id = "-" }: { id?: string }) {
  return (
    <List.Value>
      <Flex gap={3} width="full">
        <CopyButton text={id || ""} data-test-id="copy-button" />
        <Tooltip message={id || ""}>
          <Text className={styles["text-overflow"]} weight="medium">
            {id}
          </Text>
        </Tooltip>
      </Flex>
    </List.Value>
  );
}
