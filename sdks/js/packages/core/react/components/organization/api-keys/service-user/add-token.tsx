import { Dialog, Image, Separator } from '@raystack/apsara';
import styles from './styles.module.css';
import { Button, Flex, Text } from '@raystack/apsara/v1';
import cross from '~/react/assets/cross.svg';
import { useNavigate, useParams } from '@tanstack/react-router';

export default function AddServiceUserToken() {
  const { id: serviceUserId } = useParams({ from: '/api-keys/$id/add-token' });
  const navigate = useNavigate({ from: '/api-keys/$id/add-token' });

  return (
    <Dialog open={true}>
      {/* @ts-ignore */}
      <Dialog.Content
        overlayClassname={styles.overlay}
        className={styles.addDialogContent}
      >
        <Flex justify="between" className={styles.addDialogForm}>
          <Text size={6} weight={500}>
            New Api Key
          </Text>

          <Image
            alt="cross"
            style={{ cursor: 'pointer' }}
            // @ts-ignore
            src={cross}
            onClick={() =>
              navigate({
                to: '/api-keys/$id',
                params: {
                  id: serviceUserId
                }
              })
            }
            data-test-id="frontier-sdk-new-service-account-token-close-btn"
          />
        </Flex>
        <Separator />

        <Flex
          direction="column"
          gap="medium"
          className={styles.addDialogFormContent}
        >
          <Text></Text>
        </Flex>
        <Separator />
        <Flex justify="end" className={styles.addDialogFormBtnWrapper}>
          <Button
            variant="primary"
            size="normal"
            type="submit"
            data-test-id="frontier-sdk-add-service-account-token-btn"
            loaderText={'Generating...'}
          >
            Generate
          </Button>
        </Flex>
      </Dialog.Content>
    </Dialog>
  );
}
