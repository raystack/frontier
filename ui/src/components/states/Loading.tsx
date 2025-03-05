import { ClockIcon } from "@radix-ui/react-icons";
import { EmptyState, Flex } from "@raystack/apsara/v1";

export default function LoadingState() {
  return (
    <Flex style={{ height: "100vh" }}>
      <EmptyState icon={<ClockIcon />} heading="Loading...."></EmptyState>
    </Flex>
  );
}
