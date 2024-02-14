import { EmptyState, Flex, Text } from "@raystack/apsara";

export default function LoadingState() {
  return (
    <Flex style={{ height: "100vh" }}>
      <EmptyState>
        <Text size={5}>Loading....</Text>
      </EmptyState>
    </Flex>
  );
}
