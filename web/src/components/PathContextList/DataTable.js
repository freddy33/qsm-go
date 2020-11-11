import React, { useEffect, useState, useReducer } from 'react';
import { Link } from '@reach/router';
import { Table, Button } from 'semantic-ui-react';
import _ from 'lodash';

const growthTypeOptions = [1, 2, 3, 4, 8].map((v) => ({ value: v, text: v }));
const growthIndexOptions = [...Array(12).keys()].map((v) => ({ value: v, text: v }));
const growthOffsetOptions = [...Array(12).keys()].map((v) => ({ value: v, text: v }));

const DIRECTION = {
  ASCENDING: 'ascending',
  DESCENDING: 'descending',
};

const DataTable = (props) => {
  const { updateMaxDist, pathContexts } = props;

  const [direction, setDirection] = useState(DIRECTION.ASCENDING);
  const [column, setColumn] = useState();
  const [data, setData] = useState(pathContexts);

  const sort = (selectedColumn) => {
    const newDirection = direction === DIRECTION.ASCENDING ? DIRECTION.DESCENDING : DIRECTION.ASCENDING;

    const sorted = _.sortBy(data, [column]);

    if (newDirection === DIRECTION.DESCENDING) {
      setData(sorted.reverse());
    } else {
      setData(sorted);
    }

    setColumn(selectedColumn);
    setDirection(newDirection);
  };

  useEffect(() => {
    setData(pathContexts);
  }, [pathContexts]);

  return (
    <Table sortable celled>
      <Table.Header>
        <Table.Row>
          <Table.HeaderCell sorted={column === 'path_ctx_id' ? direction : null} onClick={() => sort('path_ctx_id')}>
            Path Context ID
          </Table.HeaderCell>
          <Table.HeaderCell sorted={column === 'max_dist' ? direction : null} onClick={() => sort('max_dist')}>
            Max Dist
          </Table.HeaderCell>
          <Table.HeaderCell sorted={column === 'growth_type' ? direction : null} onClick={() => sort('growth_type')}>
            Growth Type
          </Table.HeaderCell>
          <Table.HeaderCell sorted={column === 'growth_index' ? direction : null} onClick={() => sort('growth_index')}>
            Growth Index
          </Table.HeaderCell>
          <Table.HeaderCell sorted={column === 'growth_type' ? direction : null} onClick={() => sort('growth_type')}>
            Growth Offset
          </Table.HeaderCell>
          <Table.HeaderCell>Actions</Table.HeaderCell>
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {data.map((pathContext) => {
          const {
            path_ctx_id: pathContextId,
            max_dist: maxDist,
            growth_type: growthType,
            growth_index: growthIndex,
            growth_offset: growthOffset,
          } = pathContext;

          return (
            <Table.Row>
              <Table.Cell>
                <Link to={`render/${pathContextId}`}>{pathContextId}</Link>
              </Table.Cell>
              <Table.Cell>{maxDist}</Table.Cell>
              <Table.Cell>{growthType}</Table.Cell>
              <Table.Cell>{growthIndex}</Table.Cell>
              <Table.Cell>{growthOffset}</Table.Cell>
              <Table.Cell>
                <Link to={`render/${pathContextId}`}>
                  <Button>Render</Button>
                </Link>

                <Button onClick={() => updateMaxDist(pathContextId, maxDist + 1)}>Max dist + 1</Button>
              </Table.Cell>
            </Table.Row>
          );
        })}
      </Table.Body>
    </Table>
  );
};

export default DataTable;
