import { Flex, Spinner } from "@raystack/apsara";

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
