import React, { useEffect, useState } from 'react';
import { Button, Form, Message, Segment } from 'semantic-ui-react';
import _ from 'lodash';
import { useNavigate } from '@reach/router';

import DataTable from '../shared/DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';
import localStorage from '../../libs/util/localStorage';

const EnvironmentPage = ({ changeEnv }) => {
  const navigate = useNavigate();
  const [environments, setEnvironments] = useState([]);
  const [envIdToBeCreated, setEnvIdToBeCreated] = useState();
  const [displayMessage, setDisplayMessage] = useState('');

  const getEnvironments = async () => {
    const envs = await Service.getEnvironments();
    const sorted = _.sortBy(envs, ['env_id']);
    setEnvironments(sorted);
  };

  const createEnv = async (envId) => {
    await Service.createEnvironment(envId);
    await getEnvironments();
    setDisplayMessage('Environment is created successfully.');
  };

  const deleteEnv = async (envId) => {
    await Service.deleteEnvironment(envId);
    await getEnvironments();
  };

  useEffect(() => {
    getEnvironments();
  }, []);

  return (
    <div className={styles.environmentPage}>
      <Segment>
        <Form onSubmit={() => createEnv(envIdToBeCreated)}>
          <Form.Input
            placeholder="Env ID"
            onChange={(e, { value }) => setEnvIdToBeCreated(value)}
          />
          <Form.Button disabled={!envIdToBeCreated}>
            Create Environment
          </Form.Button>
        </Form>
        {displayMessage && (
          <Message positive>
            <Message.Header>{displayMessage}</Message.Header>
          </Message>
        )}
      </Segment>
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
        actionProducer={(rowData) => {
          return (
            <div>
              <Button
                onClick={() => {
                  changeEnv(rowData.envId);
                  navigate(`/path-contexts`);
                }}
              >
                Select
              </Button>
              <Button onClick={() => deleteEnv(rowData.envId)}>Delete</Button>
            </div>
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
