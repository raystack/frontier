import { Flex, Image, Text } from "@raystack/apsara";

type DialogHeaderProps = {
  title?: string;
};
export function DialogHeader({ title }: DialogHeaderProps) {
  return (
    <Flex
      justify="between"
      align="center"
      style={{ padding: "16px 32px", width: "98%", height: "52px" }}
    >
      <Text size="regular" weight="medium">{title}</Text>
      <Image alt="share-image" src="/share.svg" />
    </Flex>
  );
}
