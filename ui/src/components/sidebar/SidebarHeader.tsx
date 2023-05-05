import { Flex, Text } from "@odpf/apsara";

const BRAND_NAME = "Shield";
export default function SidebarHeader() {
  return (
    <Flex align="center" direction="row" css={sidebarHeaderContainerStyle}>
      <img src="/console/logo.svg" alt="shield" width={16} height={16} />
      <Text css={{ marginLeft: "8px", fontWeight: "500" }}>{BRAND_NAME}</Text>
    </Flex>
  );
}

const sidebarHeaderContainerStyle = {
  height: "2rem",
  margin: "10px 1rem",
  fontSize: "1.6rem",
  padding: "$2",
};
