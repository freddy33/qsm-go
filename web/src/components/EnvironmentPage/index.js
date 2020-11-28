import React, { useEffect, useState } from 'react';
import { Button } from 'semantic-ui-react';
import _ from 'lodash';
import { useNavigate } from '@reach/router';

import DataTable from '../shared/DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';
import localStorage from '../../libs/util/localStorage';

const EnvironmentPage = ({ changeEnv }) => {
  const navigate = useNavigate();
  const [environments, setEnvironments] = useState([]);

  const getEnvironments = async () => {
    const envs = await Service.getEnvironments();
    const sorted = _.sortBy(envs, ['env_id']);
    setEnvironments(sorted);
  };

  useEffect(() => {
    getEnvironments();
  }, []);

  return (
    <div className={styles.environmentPage}>
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
                changeEnv(rowData.envId);
                navigate(`/path-contexts`);
              }}
            >
              Select
            </Button>
          );
        }}
        highlightProducer={(row) =>
          _.parseInt(localStorage.getCurrentEnv()) === row.envId
        }
      />
    </div>
  );
};

export default EnvironmentPage;
