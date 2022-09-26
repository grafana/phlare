import { css, cx } from '@emotion/css';
import React, { useMemo } from 'react';
import {
  SortByFn,
  useSortBy, 
  useAbsoluteLayout, 
  useTable,
  CellProps,
} from 'react-table';
import { FixedSizeList } from 'react-window';

import { Icon, useStyles2, CustomScrollbar } from '@grafana/ui';
import { GrafanaTheme2 } from '@grafana/data';

import { ColumnTypes, TopTableData, TopTableValue } from '../types';

type Props = {
  width: number;
  height: number;
  data: TopTableData[];
  query: string;
  setQuery: (query: string) => void;
};

const FlameGraphTopTable = ({ width, height, data, query, setQuery }: Props) => {
  const styles = useStyles2((theme) => getStyles(theme));
  const COLUMN_WIDTH = 120;

  const sortSymbols: SortByFn<object> = (a, b, column) => {
    return a.values[column].localeCompare(b.values[column]);
  }

  const sortUnits: SortByFn<object> = (a, b, column) => {
    return a.values[column].value.toString().localeCompare(b.values[column].value.toString(), "en", { numeric: true })
  }

  const columns = useMemo<any>(
    () => [
      {
        accessor: ColumnTypes.Symbol.toLowerCase(),
        header: ColumnTypes.Symbol,
        cell: SymbolCell,
        sortType: sortSymbols,
        width: width - (COLUMN_WIDTH * 2),
      },
      {
        accessor: ColumnTypes.Self.toLowerCase(),
        header: ColumnTypes.Self,
        cell: UnitCell,
        sortType: sortUnits,
        width: COLUMN_WIDTH,
      },
      {
        accessor: ColumnTypes.Total.toLowerCase(),
        header: ColumnTypes.Total,
        cell: UnitCell,
        sortType: sortUnits,
        width: COLUMN_WIDTH,
      },
    ],
    [width]
  );

  const options = useMemo(
    () => ({
      columns,
      data,
      initialState: {
        sortBy: [{
          id: ColumnTypes.Self.toLowerCase(),
          desc: true
        }]
      }
    }),
    [columns, data]
  );

  const { headerGroups, rows, prepareRow } = useTable(options, useSortBy, useAbsoluteLayout);

  const renderRow = React.useCallback(
    ({ index, style }) => {
      let row = rows[index];
      prepareRow(row);

      const rowValue = row.values[ColumnTypes.Symbol.toLowerCase()];
      const classNames = cx(rowValue === query && styles.matchedRow, styles.row);
      const rowClicked = (row: string) => {
        query === row ? setQuery('') : setQuery(row);
      };

      return (
        <div {...row.getRowProps({ style })} className={classNames} onClick={() => { rowClicked(rowValue) }}>
          {row.cells.map((cell) => {
            const { key, ...cellProps } = cell.getCellProps();
            return (
              <div key={key} className={styles.cell} {...cellProps}>
                {cell.render('cell')}
              </div>
            );
          })}
        </div>
      );
    },
    [rows, prepareRow, query, styles.matchedRow, styles.row, styles.cell, setQuery]
  );

  return (
    <div className={styles.table(height)}>
      {headerGroups.map((headerGroup) => {
        const { key, ...headerGroupProps } = headerGroup.getHeaderGroupProps();

        return (
          <div key={key} className={styles.header} {...headerGroupProps}>
            {headerGroup.headers.map((column) => {
              const { key, ...headerProps } = column.getHeaderProps(
                column.canSort ? column.getSortByToggleProps() : undefined
              );

              return (
                <div key={key} className={styles.headerCell} {...headerProps}>
                  {column.render('header')}
                  {column.isSorted && <Icon name={column.isSortedDesc ? 'arrow-down' : 'arrow-up'} />}
                </div>
              );
            })}
          </div>
        );
      })}

      {rows.length > 0 ? (
        <CustomScrollbar hideVerticalTrack={true}>
          <FixedSizeList
            height={height}
            itemCount={rows.length}
            itemSize={35}
            width={'100%'}
            style={{ overflow: 'hidden auto' }}
          >
            {renderRow}
          </FixedSizeList>
        </CustomScrollbar>
      ) : (
        <div style={{ height: height }} className={styles.noData}>
          No data
        </div>
      )}
    </div>
  );
}

const SymbolCell = ({cell: { value }}: CellProps<TopTableValue, TopTableValue>) => { 
  return <span>{value}</span>
}

const UnitCell = ({cell: { value }}: CellProps<TopTableValue, TopTableValue>) => {
  return <span>{value.unitValue}</span>
}

const getStyles = (theme: GrafanaTheme2) => ({
  table: (height: number) => {
    return css`
      label: joey-table;
      height: ${height}px;
      overflow: scroll;
      display: flex;
      flex-direction: column;
      width: 100%;
    `;
  },
  header: css`
    label: joey-header;
    height: 38px;
  `,
  headerCell: css`
    label: joey-header-cell;
    background-color: ${theme.colors.background.secondary};
    padding: ${theme.spacing(1)};
  `,
  matchedRow: css`
    label: joey-matched-row;
    display: block;

    & > :nth-child(1), & > :nth-child(2), & > :nth-child(3) {
      background-color: ${theme.colors.background.secondary} !important;
    }
  `,
  row: css`
    label: joey-row;

    &:hover {
      background-color: ${theme.colors.emphasize(theme.colors.background.primary, 0.03)};
    }
  `,
  cell: css`
    label: joey-cell;
    padding: ${theme.spacing(1)};
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;

    &:hover {
      overflow: visible;
      width: auto !important;
      box-shadow: 0 0 2px ${theme.colors.primary.main};
      background-color: ${theme.colors.background.primary};
      z-index: 1;
    }
  `,
  noData: css`
    align-items: center;
    display: flex;
    justify-content: center;
  `,
});

export default FlameGraphTopTable;
