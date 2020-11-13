import React, { useEffect, useState } from 'react';
import { Link } from '@reach/router';
import { Button, Form, Message, Segment } from 'semantic-ui-react';
import _ from 'lodash';

import DataTable from './DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';

const growthTypeOptions = [1, 2, 3, 4, 8].map((v) => ({ value: v, text: v }));
const growthIndexOptions = [...Array(12).keys()].map((v) => ({ value: v, text: v }));
const growthOffsetOptions = [...Array(12).keys()].map((v) => ({ value: v, text: v }));

const PathContextList = () => {
  const [pathContexts, setPathContexts] = useState([]);
  const [growthType, setGrowthType] = useState(growthTypeOptions[0].value);
  const [growthIndex, setGrowthIndex] = useState(growthIndexOptions[0].value);
  const [growthOffset, setGrowthOffset] = useState(growthOffsetOptions[0].value);
  const [createdPathContext, setCreatedPathContext] = useState();

  const getPathContexts = async () => {
    const pathContexts = await Service.getPathContexts();
    const sorted = _.sortBy(pathContexts, ['path_ctx_id']);
    setPathContexts(sorted);
  };

  const updateMaxDist = async (pathContextId, dist) => {
    await Service.updateMaxDist(pathContextId, dist);

    await getPathContexts();
  };

  const onSubmit = async (growthType, growthIndex, growthOffset) => {
    const resp = await Service.createPathContext(growthType, growthIndex, growthOffset);
    const { path_ctx_id: pathContextId, max_dist: maxDist } = resp;
    setCreatedPathContext({
      pathContextId,
      maxDist,
    });
    await getPathContexts();
  };

  useEffect(() => {
    getPathContexts();
  }, []);

  return (
    <div className={styles.pathContextList}>
      <h1>Path Contexts</h1>
      <Segment>
        <Form onSubmit={() => onSubmit(growthType, growthIndex, growthOffset)}>
          <Form.Group widths="equal">
            <Form.Select
              fluid
              label="Growth Type"
              options={growthTypeOptions}
              placeholder="Growth Type"
              defaultValue={_.last(growthTypeOptions).value}
              onChange={(e, { value }) => setGrowthType(value)}
            />
            <Form.Select
              fluid
              label="Growth Index"
              options={growthIndexOptions}
              placeholder="Growth Index"
              defaultValue={_.first(growthIndexOptions).value}
              onChange={(e, { value }) => setGrowthIndex(value)}
            />
            <Form.Select
              fluid
              label="Growth Offset"
              options={growthOffsetOptions}
              placeholder="Growth Offset"
              defaultValue={_.first(growthOffsetOptions).value}
              onChange={(e, { value }) => setGrowthOffset(value)}
            />
          </Form.Group>
          <Form.Button>Submit</Form.Button>
        </Form>
        {createdPathContext && (
          <Message positive>
            <Message.Header>Path context is created successfully.</Message.Header>
            <p>
              Path Context ID: {createdPathContext.pathContextId}
              <br />
              Max Dist: {createdPathContext.maxDist}
            </p>
            <Link to={`render/${createdPathContext.pathContextId}`}>
              <Button>Render</Button>
            </Link>
          </Message>
        )}
      </Segment>
      <DataTable pathContexts={pathContexts} updateMaxDist={updateMaxDist} />
    </div>
  );
};

export default PathContextList;
