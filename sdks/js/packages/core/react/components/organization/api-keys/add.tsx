import { Dialog, Separator, Image } from '@raystack/apsara';
import styles from '../organization.module.css';
import { Flex, Text } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate } from '@tanstack/react-router';

export const AddServiceAccount = () => {
  const navigate = useNavigate({ from: '/api-keys/add' });
  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        style={{ padding: 0, maxWidth: '600px', width: '100%', zIndex: '60' }}
        overlayClassname={styles.overlay}
      >
        <Flex justify="between" style={{ padding: '16px 24px' }}>
          <Text size={6} style={{ fontWeight: '500' }}>
            New Service Account
          </Text>

          <Image
            alt="cross"
            style={{ cursor: 'pointer' }}
            // @ts-ignore
            src={cross}
            onClick={() => navigate({ to: '/api-keys' })}
            data-test-id="frontier-sdk-new-service-account-close-btn"
          />
        </Flex>
        <Separator />

        <Flex
          direction="column"
          gap="medium"
          style={{ padding: '24px 32px' }}
        ></Flex>
      </Dialog.Content>
    </Dialog>
  );
};
