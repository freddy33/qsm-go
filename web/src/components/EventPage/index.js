import React, { useEffect, useState } from 'react';
import { Button, Form, Message, Segment } from 'semantic-ui-react';
import _ from 'lodash';
import { Link, useNavigate } from '@reach/router';

import DataTable from '../shared/DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';

const getColorName = (color) => {
  switch (color) {
    case 1:
      return 'Red';
    case 2:
      return 'Green';
    case 3:
      return 'Blue';
    case 4:
      return 'Yellow';
    default:
      return 'Unknown';
  }
};

const EventPage = (props) => {
  const navigate = useNavigate();
  const [events, setEvents] = useState([]);

  const { spaceId } = props;

  const getEvents = async () => {
    const events = await Service.getEvents(spaceId);
    const eventsWithExtraInfo = events.map((event) => ({
      ...event,
      color_name: `${event.color} - ${getColorName(event.color)}`,
      path_context_max_dist: _.get(event, 'root_node.d', 0),
    }));
    const sorted = _.sortBy(eventsWithExtraInfo, ['event_id']);
    setEvents(sorted);
  };

  useEffect(() => {
    getEvents();
  }, []);

  return (
    <div className={styles.eventPage}>
      <DataTable
        headers={[
          { label: 'Event ID', fieldName: 'eventId' },
          { label: 'Growth Type', fieldName: 'growthType' },
          { label: 'Growth Index', fieldName: 'growthIndex' },
          { label: 'Growth Offset', fieldName: 'growthOffset' },
          { label: 'Creation Time', fieldName: 'creationTime' },
          { label: 'Color', fieldName: 'color' },
          { label: 'Path Context Max Dist', fieldName: 'pathContextMaxDist' },
        ]}
        data={events.map((event) => ({
          eventId: event.event_id,
          growthType: event.growth_type,
          growthIndex: event.growth_index,
          growthOffset: event.growth_offset,
          creationTime: event.creation_time,
          color: event.color_name,
          pathContextMaxDist: event.path_context_max_dist,
        }))}
      />
    </div>
  );
};

export default EventPage;
