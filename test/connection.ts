import { createConnection, getConnection, getManager } from 'typeorm';
import ormConfig from '../src/config/ormconfig';

const connectionWrapper = {
  async create() {
    await createConnection(ormConfig);
  },

  async close() {
    await getConnection().close();
  },

  async clear() {
    const connection = getConnection();
    const entities = connection.entityMetadatas
      .map((entityMetadata) => entityMetadata.tableName)
      .concat(['casbin_rule']);

    const deleteQuery = `TRUNCATE ${entities.join(
      ', '
    )} RESTART IDENTITY CASCADE`;
    await getManager().query(deleteQuery);
  }
};

export default connectionWrapper;
