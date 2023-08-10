import logo from '~/react/assets/logo.png';
import {
  V1Beta1Organization,
  V1Beta1User
} from '../../../client/data-contracts';
import { Flex } from '@raystack/apsara';

const styles = {
  container: {
    minWidth: '240px',
    borderRadius: '2px',
    border: '1px solid var(--border-subtle)',
    background: 'var(--background-base)',
    boxShadow: '0px 1px 4px 0px rgba(0, 0, 0, 0.09)'
  },
  orgList: {
    width: '100%',
    padding: 'var(--pd-8) var(--pd-12)'
  },
  orgListItem: {
    display: 'flex',
    justifyContent: 'space-between',
    width: '100%',
    padding: 'var(--pd-8)'
  },
  orgName: {
    fontSize: '12px',
    fontStyle: 'normal',
    fontWeight: 500,
    lineHeight: '16px',
    letterSpacing: '0.5px'
  },
  dropdownLinksList: {
    padding: '8px 12px',
    borderTop: '1px solid var(--border-subtle)'
  },
  dropdownLinksListItem: {
    color: 'var(--foreground-muted)',
    fontSize: '12px',
    fontStyle: 'normal',
    fontWeight: 500,
    lineHeight: '16px' /* 133.333% */,
    letterSpacing: '0.5px',
    display: 'flex',
    padding: 'var(--gap-3, 8px)',
    alignItems: 'center'
  },
  orgLogo: {
    height: '16px',
    width: '16px',
    marginRight: '8px'
  },
  userProfile: {
    display: 'flex',
    padding: '12px',
    flexDirection: 'row',
    alignItems: 'flex-start'
  },
  userAvatar: {
    width: '32px',
    height: '32px',
    borderRadius: '50%',
    marginRight: '8px'
  },
  userName: {
    color: 'var(--foreground-base)',
    fontSize: '12px',
    fontStyle: 'normal',
    fontWeight: 500,
    lineHeight: '16px',
    letterSpacing: '0.5px'
  },
  userEmail: {
    color: 'var(--foreground-subtle)',
    fontSize: '11px',
    fontStyle: 'normal',
    fontWeight: 500,
    lineHeight: '16px',
    letterSpacing: '0.5px'
  }
};

const OrganizationsList = ({
  organizations,
  selectedOrgId
}: {
  organizations: V1Beta1Organization[];
  selectedOrgId?: string;
}) => {
  return (
    <div style={styles.orgList}>
      {organizations.map(org => (
        <div style={styles.orgListItem} key={org.id}>
          <Flex>
            <div>
              <img src={logo} style={styles.orgLogo} />
            </div>
            <div style={styles.orgName}>{org.title}</div>
          </Flex>
          <div></div>
        </div>
      ))}
    </div>
  );
};

const OrganizationDropdownRoot = ({
  children
}: {
  children: React.ReactNode;
}) => {
  return <div style={styles.container}>{children}</div>;
};

const DropdownLinksListItem = ({
  children
}: {
  children: string | React.ReactElement;
}) => {
  return <div style={styles.dropdownLinksListItem}>{children}</div>;
};

const DropdownLinksList = ({
  children
}: {
  children: React.ReactElement[] | React.ReactElement;
}) => {
  return <div style={styles.dropdownLinksList}>{children}</div>;
};

const ProfileSection = ({ user }: { user: V1Beta1User }) => {
  return (
    <div style={styles.userProfile as React.CSSProperties}>
      <img
        src={logo}
        style={styles.userAvatar}
        alt={user.name + '-profile-picture'}
      />
      <div>
        <div style={styles.userName}>{user.title}</div>
        <div style={styles.userEmail}>{user.email}</div>
      </div>
    </div>
  );
};

export const DropdownContent = Object.assign(OrganizationDropdownRoot, {
  LinksListItem: DropdownLinksListItem,
  LinksList: DropdownLinksList,
  ProfileSection: ProfileSection,
  OrganizationsList: OrganizationsList
});
