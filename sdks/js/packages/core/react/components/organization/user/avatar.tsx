'use client';

import { Flex, Image, Text } from '@raystack/apsara';

export const GeneralProfile = () => {
  return (
    <Flex direction="column" gap="small">
      <Image
        alt="Colm Tuite"
        src="https://pbs.twimg.com/profile_images/864164353771229187/Catw6Nmh_400x400.jpg"
        width={80}
        height={80}
        style={{ borderRadius: 'var(--pd-4)' }}
      />
      <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
        Pick a logo for your profile
      </Text>
    </Flex>
  );
};
