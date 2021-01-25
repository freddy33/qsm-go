import React, { useEffect, useState } from 'react';
import { Button, Form, Message, Segment } from 'semantic-ui-react';
import _ from 'lodash';

import DataTable from '../shared/DataTable';
import styles from './index.module.scss';
import Service from '../../libs/service';
import { Link } from '@reach/router';

const growthTypeOptions = [1, 2, 3, 4, 8].map((v) => ({ value: v, text: v }));
const growthIndexOptions = [...Array(12).keys()].map((v) => ({
  value: v,
  text: v,
}));
const growthOffsetOptions = [...Array(12).keys()].map((v) => ({
  value: v,
  text: v,
}));

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

const colorOptions = [1, 2, 3, 4].map((v) => ({
  value: v,
  text: getColorName(v),
}));

const EventPage = (props) => {
  const [events, setEvents] = useState([]);
  const [eventToBeCreated, setEventToBeCreated] = useState({
    growthType: _.last(growthTypeOptions).value,
    growthIndex: _.first(growthIndexOptions).value,
    growthOffset: _.first(growthOffsetOptions).value,
    creationTime: 0,
    centerX: 0,
    centerY: 0,
    centerZ: 0,
    color: _.first(colorOptions).value,
  });
  const [displayMessage, setDisplayMessage] = useState('');

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

  const createEvent = async (event) => {
    const result = await Service.createEvent(
      spaceId,
      event.growthType,
      event.growthIndex,
      event.growthOffset,
      event.creationTime,
      event.centerX,
      event.centerY,
      event.centerZ,
      event.color,
    );

    if (result) {
      setDisplayMessage(`Event is created successfully.`);
      return getEvents();
    }
  };

  useEffect(() => {
    getEvents();
  }, []);

  return (
    <div className={styles.eventPage}>
      <Segment>
        <Form onSubmit={() => createEvent(eventToBeCreated)}>
          <Form.Group widths="equal">
            <Form.Select
              fluid
              label="Growth Type"
              options={growthTypeOptions}
              placeholder="Growth Type"
              defaultValue={_.last(growthTypeOptions).value}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  growthType: value,
                })
              }
            />
            <Form.Select
              fluid
              label="Growth Index"
              options={growthIndexOptions}
              placeholder="Growth Index"
              defaultValue={_.first(growthIndexOptions).value}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  growthIndex: value,
                })
              }
            />
            <Form.Select
              fluid
              label="Growth Offset"
              options={growthOffsetOptions}
              placeholder="Growth Offset"
              defaultValue={_.first(growthOffsetOptions).value}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  growthOffset: value,
                })
              }
            />
            <Form.Select
              fluid
              label="Color"
              options={colorOptions}
              placeholder="Color"
              defaultValue={_.first(colorOptions).value}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  color: value,
                })
              }
            />
            <Form.Input
              fluid
              type="number"
              label="Creation Time"
              placeholder="Creation Time"
              value={eventToBeCreated.creationTime}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  creationTime: value,
                })
              }
            />
            <Form.Input
              fluid
              type="number"
              label="Center X"
              placeholder="X"
              value={eventToBeCreated.centerX}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  centerX: value,
                })
              }
            />
            <Form.Input
              fluid
              type="number"
              label="Center Y"
              placeholder="Y"
              value={eventToBeCreated.centerY}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  centerY: value,
                })
              }
            />
            <Form.Input
              fluid
              type="number"
              label="Center Z"
              placeholder="Z"
              value={eventToBeCreated.centerZ}
              onChange={(e, { value }) =>
                setEventToBeCreated({
                  ...eventToBeCreated,
                  centerZ: value,
                })
              }
            />
          </Form.Group>
          <Form.Button
            disabled={_.values(eventToBeCreated).some((value) => value === '')}
          >
            Create Event
          </Form.Button>
        </Form>
        {displayMessage && (
          <Message positive>
            <Message.Header>{displayMessage}</Message.Header>
          </Message>
        )}
      </Segment>
      <Segment>
        <Link to={`/render/space/${spaceId}`}>
          <Button>Render</Button>
        </Link>
      </Segment>
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
