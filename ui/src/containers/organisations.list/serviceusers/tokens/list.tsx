import { Table, Text } from "@raystack/apsara";
import styles from "./tokens.module.css";
import { useCallback, useEffect, useState } from "react";
import { V1Beta1ServiceUserToken } from "@raystack/frontier";
import { useFrontier } from "@raystack/frontier/react";
import dayjs from "dayjs";
import { DEFAULT_DATE_FORMAT } from "~/utils/constants";
import Skeleton from "react-loading-skeleton";
import { TrashIcon } from "@radix-ui/react-icons";
import { toast } from "sonner";

interface TokensListProps {
  serviceUserId: string;
}

interface TableLoaderProps {
  row?: number;
  cell?: number;
  cellClassName?: string;
}

function TableLoader({
  row = 5,
  cell = 3,
  cellClassName = "",
}: TableLoaderProps) {
  return (
    <>
      {[...new Array(row)].map((_, i) => (
        <Table.Row key={i}>
          {[...new Array(cell)].map((_, j) => (
            <Table.Cell className={cellClassName} key={i + "-" + j}>
              <Skeleton />
            </Table.Cell>
          ))}
        </Table.Row>
      ))}
    </>
  );
}

export default function TokensList({ serviceUserId }: TokensListProps) {
  const { client } = useFrontier();
  const [tokens, setTokens] = useState<V1Beta1ServiceUserToken[]>([]);
  const [isTokensLoading, setIsTokensLoading] = useState(false);

  const fetchTokens = useCallback(
    async (userId: string) => {
      try {
        setIsTokensLoading(true);
        const resp = await client?.frontierServiceListServiceUserTokens(userId);
        const tokenList = resp?.data?.tokens || [];
        setTokens(tokenList);
      } catch (err) {
        console.error(err);
      } finally {
        setIsTokensLoading(false);
      }
    },
    [client]
  );

  useEffect(() => {
    if (serviceUserId) {
      fetchTokens(serviceUserId);
    }
  }, [serviceUserId, client, fetchTokens]);

  async function deleteToken(tokenId: string) {
    const resp = await client?.frontierServiceDeleteServiceUserToken(
      serviceUserId,
      tokenId
    );

    if (resp?.status === 200) {
      toast.success("Token Deleted");
      await fetchTokens(serviceUserId);
    }
  }

  return (
    <div className={styles.tokensListWrapper}>
      <Text size={3} weight={500}>
        Tokens
      </Text>
      <Table className={styles.tokensTable}>
        <Table.Header>
          <Table.Row>
            <Table.Head className={styles.tableCell}>ID</Table.Head>
            <Table.Head className={styles.tableCell}>Title</Table.Head>
            <Table.Head className={styles.tableCell}>Created at</Table.Head>
            <Table.Head className={styles.tableCell}>Action</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {isTokensLoading ? (
            <TableLoader cell={4} cellClassName={styles.tableCell} />
          ) : (
            tokens.map((token) => (
              <Table.Row key={token.id}>
                <Table.Cell className={styles.tableCell}>{token.id}</Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  {token.title}
                </Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  {dayjs(token?.created_at).format(DEFAULT_DATE_FORMAT)}
                </Table.Cell>
                <Table.Cell className={styles.tableCell}>
                  <TrashIcon
                    className={styles.deleteIcon}
                    onClick={() => deleteToken(token?.id || "")}
                  />
                </Table.Cell>
              </Table.Row>
            ))
          )}
        </Table.Body>
      </Table>
    </div>
  );
}
