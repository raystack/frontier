import { Cross1Icon, PlusIcon } from "@radix-ui/react-icons";

import { Button, Flex, Table, useTable } from "@odpf/apsara";
import { styles } from "~/styles";

export const OrganizationsHeader = () => {
  const { filterQuery = [], clearFilters } = useTable();

  return (
    <>
      <Flex
        align="center"
        justify="between"
        css={{ width: "100%", padding: "$4 24px", fontSize: 12 }}
      >
        <Flex>Organisations</Flex>
        <Flex align="center" direction="row" css={{ columnGap: "8px" }}>
          {filterQuery.length ? (
            <Button onClick={clearFilters} css={styles.button}>
              <Flex align="center" css={{ paddingRight: "$2" }}>
                Clear Filters
              </Flex>
              <Cross1Icon />
            </Button>
          ) : (
            <Table.ColumnFilterSelection align="end">
              <Button css={styles.button}>
                <Flex align="center" css={{ paddingRight: "$2" }}>
                  <PlusIcon />
                </Flex>
                Filter
              </Button>
            </Table.ColumnFilterSelection>
          )}
          <Table.TableColumnsFilter>
            <Button css={styles.button}>View</Button>
          </Table.TableColumnsFilter>
          <Table.TableGlobalSearch placeholder="Search all organisations" />

          <Button css={styles.button}>
            <Flex align="center" css={{ paddingRight: "$2" }}>
              <PlusIcon />
            </Flex>
            new organisation
          </Button>
        </Flex>
      </Flex>
    </>
  );
};
