import React, { useEffect, useState } from 'react';
// import { Link } from '@reach/router';
import { Button } from 'semantic-ui-react';
import _ from 'lodash';

import DataTable from '../shared/DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';
import LocalStorage from '../../libs/util/localStorage';
import { LOCAL_STORAGE_KEY } from '../../libs/constant';

const EnvironmentPage = () => {
  const [environments, setEnvironments] = useState([]);

  const getEnvironments = async () => {
    const envs = await Service.getEnvironments();
    const sorted = _.sortBy(envs, ['env_id']);
    setEnvironments(sorted);
  };

  const selectEnv = (envId) => LocalStorage.setItem(LOCAL_STORAGE_KEY.SELECTED_ENVIRONMENT, envId);
  const getCurrentEnv = () => _.parseInt(LocalStorage.getItem(LOCAL_STORAGE_KEY.SELECTED_ENVIRONMENT));

  useEffect(() => {
    getEnvironments();
  }, []);

  return (
    <div className={styles.environmentPage}>
      <h1>Environments</h1>
      <DataTable
        headers={[
          { label: 'Env ID', fieldName: 'envId' },
          { label: 'Name', fieldName: 'schemaName' },
          { label: 'Size', fieldName: 'schemaSize' },
          { label: 'Size Percent', fieldName: 'schemaSizePercent' },
        ]}
        data={environments.map((env) => ({
          envId: env.env_id,
          schemaName: env.schema_name,
          schemaSize: env.schema_size,
          schemaSizePercent: env.schema_size_percent,
        }))}
        actionProducer={(rowData, rerender) => {
          return (
            <Button
              onClick={() => {
                selectEnv(rowData.envId);
                rerender();
              }}
            >
              Select
            </Button>
          );
        }}
        highlightProducer={(row) => getCurrentEnv() === row.envId}
      />
    </div>
  );
};

export default EnvironmentPage;
