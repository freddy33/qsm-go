import React, { useEffect, useState } from 'react';
import { Link } from '@reach/router';
import { Table } from 'semantic-ui-react';
import _ from 'lodash';

import styles from './PathContextList.module.scss';
import Service from '../libs/service';

const PathContextList = () => {
  const [pathContexts, setPathContexts] = useState([]);

  useEffect(() => {
    Service.getPathContexts().then((pathContexts) => {
      const sorted = _.sortBy(pathContexts, ['path_ctx_id']);
      setPathContexts(sorted);
    });
  }, []);

  return (
    <div className={styles.pathContextList}>
      <h1>Path Contexts</h1>
      <Table celled>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Path Context ID</Table.HeaderCell>
            <Table.HeaderCell>Max Dist</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {pathContexts.map((pathContext) => {
            const { path_ctx_id, max_dist } = pathContext;
            return (
              <Table.Row>
                <Table.Cell>
                  <Link to={`render/${path_ctx_id}`}>{path_ctx_id}</Link>
                </Table.Cell>
                <Table.Cell>{max_dist}</Table.Cell>
              </Table.Row>
            );
          })}
        </Table.Body>
      </Table>
    </div>
  );
};

export default PathContextList;
