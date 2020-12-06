import React, { useEffect, useState } from 'react';
import { Button, Form, Message, Segment } from 'semantic-ui-react';
import _ from 'lodash';
import { Link, useNavigate } from '@reach/router';

import DataTable from '../shared/DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';
import localStorage from '../../libs/util/localStorage';

const SpacePage = () => {
  const navigate = useNavigate();
  const [spaces, setSpaces] = useState([]);
  const [spaceToBeCreated, setSpaceToBeCreated] = useState({
    spaceName: '',
    activeThreshold: '',
    maxTime: '',
    maxCoord: '',
  });
  const [displayMessage, setDisplayMessage] = useState('');

  const getSpaces = async () => {
    const spaces = await Service.getSpaces();
    const spacesWithEventLength = spaces.map((space) => ({
      ...space,
      event_length: _.get(space, 'event_ids.length', 0),
    }));
    const sorted = _.sortBy(spacesWithEventLength, ['space_id']);
    setSpaces(sorted);
  };

  const createSpace = async (space) => {
    const { spaceName, activeThreshold, maxTime, maxCoord } = space;
    const result = await Service.createSpace(
      spaceName,
      activeThreshold,
      maxTime,
      maxCoord,
    );

    if (result) {
      setDisplayMessage(`Space ${spaceName} is created successfully.`);
      return getSpaces();
    }
  };

  const deleteSpace = async (spaceId, spaceName) => {
    const result = await Service.deleteSpace(spaceId, spaceName);

    if (result) {
      await getSpaces();
    }
  };

  useEffect(() => {
    getSpaces();
  }, []);

  return (
    <div className={styles.spacePage}>
      <Segment>
        <Form onSubmit={() => createSpace(spaceToBeCreated)}>
          <Form.Input
            placeholder="Space Name"
            onChange={(e, { value }) =>
              setSpaceToBeCreated({
                ...spaceToBeCreated,
                spaceName: value,
              })
            }
          />
          <Form.Input
            type="number"
            placeholder="Active Threshold"
            onChange={(e, { value }) =>
              setSpaceToBeCreated({
                ...spaceToBeCreated,
                activeThreshold: value,
              })
            }
          />
          <Form.Input
            type="number"
            placeholder="Max Time"
            onChange={(e, { value }) =>
              setSpaceToBeCreated({
                ...spaceToBeCreated,
                maxTime: value,
              })
            }
          />
          <Form.Input
            type="number"
            placeholder="Max Coord"
            onChange={(e, { value }) =>
              setSpaceToBeCreated({
                ...spaceToBeCreated,
                maxCoord: value,
              })
            }
          />
          <Form.Button
            disabled={_.values(spaceToBeCreated).some((value) => value === '')}
          >
            Create Space
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
          { label: 'Space ID', fieldName: 'spaceId' },
          { label: 'Name', fieldName: 'spaceName' },
          { label: 'Active Threshold', fieldName: 'activeThreshold' },
          { label: 'NB Events', fieldName: 'eventLength' },
          { label: 'Max Time', fieldName: 'maxTime' },
          { label: 'Max Coord', fieldName: 'maxCoord' },
        ]}
        data={spaces.map((space) => ({
          spaceId: space.space_id,
          spaceName: space.space_name,
          activeThreshold: space.active_threshold,
          eventLength: space.event_length,
          maxTime: space.max_time,
          maxCoord: space.max_coord,
        }))}
        actionProducer={({ spaceId, spaceName }) => (
          <>
            <Link to={`/events/${spaceId}`}>
              <Button>Events</Button>
            </Link>
            <Button onClick={() => deleteSpace(spaceId, spaceName)}>
              Delete
            </Button>
          </>
        )}
      />
    </div>
  );
};

export default SpacePage;
