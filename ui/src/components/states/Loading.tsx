import { Flex, Spinner } from "@raystack/apsara/v1";

export default function LoadingState() {
  return (
    <Flex
      style={{ height: "100vh", width: "100%" }}
      justify={"center"}
      align={"center"}
    >
      <Spinner size={6} />
    </Flex>
  );
}
