import { css } from '@emotion/css';
import React, { useCallback, useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import { useStyles2 } from '@grafana/ui';

import { PIXELS_PER_LEVEL } from '../../constants';
import { Item } from '../FlameGraph/dataTransform';
import { SampleUnit, SelectedView, TableData, TopTableData } from '../types';
import FlameGraphTopTable from './FlameGraphTopTable';
import { createTheme, DataFrame, Field, FieldType, getDisplayProcessor } from '@grafana/data';

type Props = {
  data: DataFrame;
  levels: Item[][];
  selectedView: SelectedView;
  query: string;
  setQuery: (query: string) => void;
};

const FlameGraphTopTableContainer = ({ data, levels, selectedView, query, setQuery }: Props) => {
  const styles = useStyles2(() => getStyles(selectedView));
  const [topTable, setTopTable] = useState<TopTableData[]>();
  const valueField =
    data.fields.find((f) => f.name === 'value') ?? data.fields.find((f) => f.type === FieldType.number);
  const selfField =
    data.fields.find((f) => f.name === 'self') ?? data.fields.find((f) => f.type === FieldType.number);

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

  const getTopTableData = (field: Field, value: number) => {
    const processor = getDisplayProcessor({ field, theme: createTheme() /* theme does not matter for us here */ });
    const displayValue = processor(value);
    let unitValue = displayValue.text + displayValue.suffix;
  
    switch (field.config.unit) {
      case SampleUnit.Bytes:
        break;
      case SampleUnit.Nanoseconds:
        break;
      default:
        if (!displayValue.suffix) {
          // Makes sure we don't show 123undefined or something like that if suffix isn't defined
          unitValue = displayValue.text
        }
        break;
    }
  
    return unitValue;
  };

  useEffect(() => {
    const table = sortLevelsIntoTable();

    let topTable: TopTableData[] = [];
    for (let key in table) {
      const selfUnit = getTopTableData(selfField!, table[key].self);
      const valueUnit = getTopTableData(valueField!, table[key].total);

      topTable.push({
        symbol: key,
        self: { value: table[key].self, unitValue: selfUnit },
        total: { value: table[key].total, unitValue: valueUnit }
      });
    }

    setTopTable(topTable);
  }, [data.fields, selfField, sortLevelsIntoTable, valueField]);


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
