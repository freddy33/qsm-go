import React, { useEffect, useState } from 'react';
import { Table } from 'semantic-ui-react';
import _ from 'lodash';

const DIRECTION = {
  ASCENDING: 'ascending',
  DESCENDING: 'descending',
};

const Index = (props) => {
  const { headers, data, actionProducer, highlightProducer } = props;

  const [direction, setDirection] = useState(DIRECTION.ASCENDING);
  const [sortedColumn, setSortedColumn] = useState();
  const [currentData, setCurrentData] = useState(data);
  const [lastRender, setLastRender] = useState(Date.now());

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

  const rerender = () => setLastRender(Date.now());

  useEffect(() => {
    setCurrentData(data);
  }, [data]);

  useEffect(() => {
    // use this state to force UI to rerender
  }, [lastRender]);

  const shouldShowAction = !!actionProducer;
  return (
    <Table sortable celled>
      <Table.Header>
        <Table.Row>
          {headers.map((header) => {
            const { fieldName, label, sortable = true } = header;
            const cellProperties = sortable
              ? {
                  sorted: sortedColumn === fieldName ? direction : null,
                  onClick: () => sort(fieldName),
                }
              : {};
            return <Table.HeaderCell {...cellProperties}>{label}</Table.HeaderCell>;
          })}
          {shouldShowAction && <Table.HeaderCell>Actions</Table.HeaderCell>}
        </Table.Row>
      </Table.Header>

      <Table.Body>
        {currentData.map((row) => (
          <Table.Row positive={highlightProducer && highlightProducer(row)}>
            {headers.map(({ fieldName }) => (
              <Table.Cell>{row[fieldName]}</Table.Cell>
            ))}
            {shouldShowAction && <Table.Cell>{actionProducer(row, rerender)}</Table.Cell>}
          </Table.Row>
        ))}
      </Table.Body>
    </Table>
  );
};

export default Index;
