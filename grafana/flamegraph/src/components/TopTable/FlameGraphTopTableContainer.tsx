import { css } from '@emotion/css';
import React, { useCallback, useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import { useStyles2 } from '@grafana/ui';

import { PIXELS_PER_LEVEL } from '../../constants';
import { Item } from '../FlameGraph/dataTransform';
import { SelectedView, TableData, TopTableData } from '../types';
import FlameGraphTopTable from './FlameGraphTopTable';

type Props = {
  levels: Item[][];
  profileTypeId: string;
  selectedView: SelectedView;
  query: string;
  setQuery: (query: string) => void;
};

const FlameGraphTopTableContainer = ({ levels, profileTypeId, selectedView, query, setQuery }: Props) => {
  const styles = useStyles2(() => getStyles(selectedView));
  const [topTable, setTopTable] = useState<TopTableData[]>();

  const sortLevelsIntoTable = useCallback(() => {
    let label, self, value;
    let item: Item;
    let table: { [key: string]: TableData } = {};

    for (let i = 0; i < levels.length; i++) {
      for (var j = 0; j < Object.values(levels[i]).length; j++) {
        item = Object.values(levels[i])[j];
        label = item.label;
        self = item.self;
        value = item.value;
        table[label] = table[label] || {};
        table[label].self = table[label].self ? table[label].self + self : self;
        table[label].total = table[label].total ? table[label].total + value : value;
      }
    }

    return table;
  }, [levels]);

  useEffect(() => {
    const table = sortLevelsIntoTable();

    let topTable: TopTableData[] = [];
    for (let key in table) {
      topTable.push({
        symbol: key,
        // self: { value: table[key].self, unitValue: getUnitValue(table[key].self / divisor, unit, fallback) },
        // total: { value: table[key].total, unitValue: getUnitValue(table[key].total / divisor, unit, fallback) }
        self: { value: table[key].self, unitValue: '1010100101' },
        total: { value: table[key].total, unitValue: '1010100101' }
      });
    }

    setTopTable(topTable);
  }, [sortLevelsIntoTable]);


  return (
    <>
      {topTable && (
        <div className={styles.topTableContainer}>
          <AutoSizer style={{ width: '100%', height: PIXELS_PER_LEVEL * levels.length + 'px' }}>
            {({ width, height }) => (
              <FlameGraphTopTable width={width} height={height} data={topTable} query={query} setQuery={setQuery} />
            )}
          </AutoSizer>
        </div>
      )}
    </>
  );
};

const getStyles = (selectedView: SelectedView) => {
  const marginRight = '20px';

  return {
    topTableContainer: css`
      cursor: pointer;
      float: left;
      margin-right: ${marginRight};
      width: ${selectedView === SelectedView.TopTable ? '100%' : `calc(50% - ${marginRight})`};
    `,
  }
};

export default FlameGraphTopTableContainer;
