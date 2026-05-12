import { CopyButton, Flex, List, Text, Tooltip } from "@raystack/apsara-v1";
import styles from "./audit-logs.module.css";

export default function SidepanelListId({ id = "-" }: { id?: string }) {
  return (
    <List.Value>
      <Flex gap={3} width="full">
        <CopyButton text={id || ""} data-test-id="copy-button" />
        <Tooltip>
          <Tooltip.Trigger
            render={
              <Text className={styles["text-overflow"]} weight="medium">
                {id}
              </Text>
            }
          />
          <Tooltip.Content>{id || ""}</Tooltip.Content>
        </Tooltip>
      </Flex>
    </List.Value>
  );
}
