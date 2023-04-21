import { Cross1Icon, PlusIcon } from "@radix-ui/react-icons";

import { Button, Flex, Table, Text, useTable } from "@odpf/apsara";
import { useNavigate } from "react-router-dom";
import { styles } from "~/styles";

export const UsersHeader = () => {
  const navigate = useNavigate();
  const { filterQuery = [], clearFilters } = useTable();

  return (
    <>
      <Flex
        align="center"
        justify="between"
        css={{ width: "100%", padding: "$4 24px", fontSize: 12 }}
      >
        <Text size={4} css={{ fontWeight: "500" }}>
          Users
        </Text>
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
              <Button css={{ ...styles.button, border: "1px dashed $gray6" }}>
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
          <Table.TableGlobalSearch
            css={{ height: "24px" }}
            placeholder="Search all users"
          />

          <Button css={styles.button} onClick={() => navigate("/users/create")}>
            <Flex align="center" css={{ paddingRight: "$2" }}>
              <PlusIcon />
            </Flex>
            New user
          </Button>
        </Flex>
      </Flex>
    </>
  );
};
