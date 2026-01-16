"use client";

import { Box, Flex, Image } from '@raystack/apsara';
import { Header } from '../../../react/components/Header';
import { MagicLink } from '../../../react/components/onboarding/magiclink';
import { AdminPageTitle } from '../../components/PageTitle';
import type { AdminLoginProps } from '../../types';

export function AdminLogin({ config, logoIcon }: AdminLoginProps) {
  return (
    <Flex>
      <AdminPageTitle title="Login" config={config} />
      <Box style={{ width: '100%' }}>
        <Flex
          direction="column"
          justify="center"
          align="center"
          style={{
            margin: 'auto',
            height: '100vh',
            width: '280px'
          }}
        >
          <Flex direction="column" gap={5} style={{ width: '100%' }}>
            <Header
              logo={
                config?.logo ? (
                  <Image
                    alt="logo"
                    src={config.logo}
                    width={80}
                    height={80}
                    style={{ borderRadius: 'var(--rs-space-3)' }}
                  />
                ) : (
                  logoIcon
                )
              }
              title={`Login to ${config?.title || 'Frontier Admin'}`}
            />
            <MagicLink open />
          </Flex>
        </Flex>
      </Box>
    </Flex>
  );
}

