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

  const getSpaces = async () => {
    const spaces = await Service.getSpaces();
    const spacesWithEventLength = spaces.map((space) => ({
      ...space,
      event_length: _.get(space, 'event_ids.length', 0),
    }));
    const sorted = _.sortBy(spacesWithEventLength, ['space_id']);
    setSpaces(sorted);
  };

  useEffect(() => {
    getSpaces();
  }, []);

  return (
    <div className={styles.spacePage}>
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
        actionProducer={({ spaceId }) => (
          <>
            <Link to={`/spaces/${spaceId}`}>
              <Button>Detail</Button>
            </Link>
          </>
        )}
      />
    </div>
  );
};

export default SpacePage;
