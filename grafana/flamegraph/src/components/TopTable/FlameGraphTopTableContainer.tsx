import { css } from '@emotion/css';
import React, { useCallback, useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import { useStyles2 } from '@grafana/ui';

import { PIXELS_PER_LEVEL } from '../../constants';
import { Item } from '../FlameGraph/dataTransform';
import { getUnitValue } from '../FlameGraph/FlameGraphTooltip';
import { BYTE_UNITS, COUNT_UNITS, NANOSECOND_UNITS, SampleUnit, SelectedView, TableData, TopTableData } from '../types';
import FlameGraphTopTable from './FlameGraphTopTable';

type Props = {
  levels: Item[][];
  profileTypeId: string;
  selectedView: SelectedView;
  query: string;
  setQuery: (query: string) => void;
};

const FlameGraphTopTableContainer = ({ levels, profileTypeId, selectedView, query, setQuery }: Props) => {
  const styles = useStyles2((theme) => getStyles(selectedView));
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

  const getUnitMeta = useCallback(() => {
    const sampleUnit = profileTypeId?.split(':').length === 5 ? profileTypeId.split(':')[2] : '';
    switch (sampleUnit) {
      case SampleUnit.Bytes:
        return { divisor: 1, unit: BYTE_UNITS, fallback: '' }
      case SampleUnit.Nanoseconds:
        return { divisor: 1000000000, unit: NANOSECOND_UNITS, fallback: 'seconds' }
      default:
        return { divisor: 1, unit: COUNT_UNITS, fallback: '' }
    }
  }, [profileTypeId]);

  useEffect(() => {
    const table = sortLevelsIntoTable();
    const { divisor, unit, fallback } = getUnitMeta();

    let topTable: TopTableData[] = [];
    for (let key in table) {
      topTable.push({
        symbol: key,
        self: { value: table[key].self, unitValue: getUnitValue(table[key].self / divisor, unit, fallback) },
        total: { value: table[key].total, unitValue: getUnitValue(table[key].total / divisor, unit, fallback) }
      });
    }

    setTopTable(topTable);
  }, [getUnitMeta, sortLevelsIntoTable]);


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

const getStyles = (selectedView: SelectedView) => ({
  topTableContainer: css`
    cursor: pointer;
    float: left;
    margin-right: 20px;
    width: ${selectedView === SelectedView.TopTable ? '100%' : 'calc(50% - 20px)'};
  `,
});

export default FlameGraphTopTableContainer;
