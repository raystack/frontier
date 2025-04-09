'use client';

import { Flex, InputField, Image, Text, Button } from '@raystack/apsara/v1';
import styles from './onboarding.module.css';
import { ReactNode } from '@tanstack/react-router';
import PixxelLogoMonogram from '~/react/assets/logos/pixxel-logo-monogram.svg';

type SubscribeProps = {
  logo?: ReactNode;
  preferenceTitle?: string;
  preferenceDescription?: string;
  onSubmit?: (data: FormData) => void;
};

export const Subscribe = ({
  logo,
  preferenceTitle = 'Updates, News & Events',
  preferenceDescription = 'Stay informed on new features, improvements, and key updates',
  onSubmit
}: SubscribeProps) => {
  
  return (
    <Flex direction="column" gap="large" align="center" justify="center">
      <Image alt="" width={88} height={88} src={PixxelLogoMonogram as unknown as string} />
      <form onSubmit={() => null}>
        <Flex
          className={styles.subscribeContainer}
          direction='column'
          justify='start'
          align="start"
          gap="large"
        >
          <Flex direction="column" gap="small">
            <Text size={6} className={styles.subscribeTitle}>{preferenceTitle}</Text>
            <Text size={4} className={styles.subscribeDesc}>{preferenceDescription}</Text>
          </Flex>

          <InputField label='Name' />
          <InputField label='Email' />
          <InputField label='Contact Number' helperText='Add country code at the start' />
          <Button type="submit" width="100%" data-test-id="sdk-demo-subscribe-button">Subscribe</Button>
        </Flex>
      </form>
    </Flex>
  );
};
