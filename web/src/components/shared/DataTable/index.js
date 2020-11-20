import React, { useEffect, useState } from 'react';
import { Table } from 'semantic-ui-react';
import _ from 'lodash';

const DIRECTION = {
  ASCENDING: 'ascending',
  DESCENDING: 'descending',
};

const Index = (props) => {
  const { headers, data } = props;

  const [direction, setDirection] = useState(DIRECTION.ASCENDING);
  const [sortedColumn, setSortedColumn] = useState();
  const [currentData, setCurrentData] = useState(data);

  const sort = (selectedColumn) => {
    const newDirection = direction === DIRECTION.ASCENDING ? DIRECTION.DESCENDING : DIRECTION.ASCENDING;

    const sorted = _.sortBy(currentData, [sortedColumn]);

    if (newDirection === DIRECTION.DESCENDING) {
      setCurrentData(sorted.reverse());
    } else {
      setCurrentData(sorted);
    }

    setSortedColumn(selectedColumn);
    setDirection(newDirection);
  };

  useEffect(() => {
    setCurrentData(data);
  }, [data]);

  return (
    <Table sortable celled>
      <Table.Header>
        <Table.Row>
          {headers.map((header, index) => {
            const { label, sortable = true } = header;
            const cellProperties = sortable
              ? {
                  sorted: sortedColumn === index ? direction : null,
                  onClick: () => sort(index),
                }
              : {};
            return <Table.HeaderCell {...cellProperties}>{label}</Table.HeaderCell>;
          })}
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {currentData.map((row) => {
          return (
            <Table.Row>
              {row.map((column) => (
                <Table.Cell>{column}</Table.Cell>
              ))}
            </Table.Row>
          );
        })}
      </Table.Body>
    </Table>
  );
};

export default Index;
