import { OrganizationsDetailsNavabar } from "./navbar";
import styles from "./details.module.css";

export const OrganizationDetails = () => {
  return (
    <div className={styles.page}>
      <OrganizationsDetailsNavabar />
      <p>This is the details page for an organization.</p>
    </div>
  );
};
