import { ReactNode, useState } from "react";
import { Flex } from "@raystack/apsara";
import { UserDetailsSidePanel } from "./side-panel";
import styles from "./layout.module.css";
import { UserDetailsNavbar } from "./navbar";

interface UserDetailsLayoutProps {
  children: ReactNode;
}

export const UserDetailsLayout = ({ children }: UserDetailsLayoutProps) => {
  const [showSidePanel, setShowSidePanel] = useState(true);

  function toggleSidePanel() {
    setShowSidePanel(!showSidePanel);
  }

  return (
    <Flex direction="column" className={styles.page}>
      <UserDetailsNavbar toggleSidePanel={toggleSidePanel} />
      <Flex justify="between" style={{ height: "100%" }}>
        <Flex
          className={
            showSidePanel
              ? styles["main_content_with_sidepanel"]
              : styles["main_content"]
          }
        >
          {children}
        </Flex>
        {showSidePanel ? <UserDetailsSidePanel /> : null}
      </Flex>
    </Flex>
  );
};
