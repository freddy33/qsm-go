import React, { useEffect, useState } from 'react';
import { Link } from '@reach/router';
import { Table, Button } from 'semantic-ui-react';
import _ from 'lodash';

import styles from './PathContextList.module.scss';
import Service from '../libs/service';

const getPathContexts = async (setPathContexts) => {
  Service.getPathContexts().then((pathContexts) => {
    const sorted = _.sortBy(pathContexts, ['path_ctx_id']);
    setPathContexts(sorted);
  });
};

const updateMaxDist = async (setPathContexts, pathContextId, dist) => {
  await Service.updateMaxDist(pathContextId, dist);

  await getPathContexts(setPathContexts);
};

const PathContextList = () => {
  const [pathContexts, setPathContexts] = useState([]);

  useEffect(() => {
    getPathContexts(setPathContexts);
  }, []);

  return (
    <div className={styles.pathContextList}>
      <h1>Path Contexts</h1>
      <Table celled>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>Path Context ID</Table.HeaderCell>
            <Table.HeaderCell>Max Dist</Table.HeaderCell>
            <Table.HeaderCell>Actions</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {pathContexts.map((pathContext) => {
            const { path_ctx_id: pathContextId, max_dist: maxDist } = pathContext;
            return (
              <Table.Row>
                <Table.Cell>
                  <Link to={`render/${pathContextId}`}>{pathContextId}</Link>
                </Table.Cell>
                <Table.Cell>{maxDist}</Table.Cell>
                <Table.Cell>
                  <Link to={`render/${pathContextId}`}>
                    <Button>Render</Button>
                  </Link>

                  <Button onClick={() => updateMaxDist(setPathContexts, pathContextId, maxDist + 1)}>
                    Increment max dist
                  </Button>
                </Table.Cell>
              </Table.Row>
            );
          })}
        </Table.Body>
      </Table>
    </div>
  );
};

export default PathContextList;
