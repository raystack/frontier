import { Flex, Text } from '@raystack/apsara';
import { styles } from '../styles';
import Skeleton from 'react-loading-skeleton';
import tokenStyles from './token.module.css';
import { useFrontier } from '~/react/contexts/FrontierContext';

interface TokenHeaderProps {
  billingSupportEmail?: string;
  isLoading?: boolean;
}

const TokensHeader = ({ billingSupportEmail, isLoading }: TokenHeaderProps) => {
  return (
    <Flex direction="column" gap="small">
      {isLoading ? (
        <Skeleton containerClassName={tokenStyles.flex1} />
      ) : (
        <Text size={6}>Tokens</Text>
      )}
      {isLoading ? (
        <Skeleton containerClassName={tokenStyles.flex1} />
      ) : (
        <Text size={4} style={{ color: 'var(--foreground-muted)' }}>
          Oversee your billing and invoices.
          {billingSupportEmail ? (
            <>
              {' '}
              For more details, contact{' '}
              <a
                href={`mailto:${billingSupportEmail}`}
                target="_blank"
                style={{ fontWeight: 400, color: 'var(--foreground-accent)' }}
              >
                {billingSupportEmail}
              </a>
            </>
          ) : null}
        </Text>
      )}
    </Flex>
  );
};

export default function Tokens() {
  const { config } = useFrontier();
  return (
    <Flex direction="column" style={{ width: '100%' }}>
      <Flex style={styles.header}>
        <Text size={6}>Tokens</Text>
      </Flex>
      <Flex direction="column" gap="large" style={styles.container}>
        <Flex direction="column" style={{ gap: '24px' }}>
          <TokensHeader billingSupportEmail={config.billing?.supportEmail} />
        </Flex>
      </Flex>
    </Flex>
  );
}
