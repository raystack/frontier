import { Flex, Text } from "@odpf/apsara";

const BRAND_NAME = "Shield";
export default function SidebarHeader() {
  return (
    <Flex align="center" direction="row" css={sidebarHeaderContainerStyle}>
      <img src="/logo.svg" alt="shield" />
      <Text css={{ marginLeft: "8px" }}>{BRAND_NAME}</Text>
    </Flex>
  );
}

const sidebarHeaderContainerStyle = {
  height: "2rem",
  fontSize: "1.6rem",
  fontWeight: "bold",
  padding: "$2",
  marginBottom: "22px",
};
