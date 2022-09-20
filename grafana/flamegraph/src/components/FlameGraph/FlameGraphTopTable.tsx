import { css } from '@emotion/css';
import React, { useCallback, useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import { Table } from '@grafana/ui';
import { ArrayVector, DataFrame, FieldType, getRawDisplayProcessor } from '@grafana/data';

import { PIXELS_PER_LEVEL } from '../../constants';
import { Item } from './dataTransform';
import { getUnitValue } from './FlameGraphTooltip';
import { BYTE_UNITS, COUNT_UNITS, NANOSECOND_UNITS, SampleUnit, SelectedView } from '../types';

type Props = {
  levels: Item[][];
  profileTypeId: string;
  selectedView: SelectedView;
};

const FlameGraphTopTable = ({ levels, profileTypeId, selectedView }: Props) => {
  const styles = getStyles(selectedView);
  const [df, setDf] = useState<DataFrame>({ fields: [], length: 0 });

  const sortLevelsIntoTable = useCallback(() => {
    let label, self, value;
    let item: Item;
    let topTable: { [key: string]: any } = [];

    for (let i = 0; i < levels.length; i++) {
      for (var j = 0; j < Object.values(levels[i]).length; j++) {
        item = Object.values(levels[i])[j];
        label = item.label;
        self = item.self;
        value = item.value;
        topTable[label] = topTable[label] || {};
        topTable[label].self = topTable[label].self ? topTable[label].self + self : self;
        topTable[label].value = topTable[label].value ? topTable[label].value + value : value;
      }
    }

    return topTable;
  }, [levels]);

  const getDisplay = useCallback(() => {
    let display;

    const sampleUnit = profileTypeId?.split(':').length === 5 ? profileTypeId.split(':')[2] : '';
    switch (sampleUnit) {
      case SampleUnit.Bytes:
        display = (v: number) => ({ numeric: v, text: getUnitValue(v, BYTE_UNITS) });
        break;

      case SampleUnit.Count:
        display = (v: number) => ({ numeric: v, text: getUnitValue(v, COUNT_UNITS) });
        break;

      case SampleUnit.Nanoseconds:
        display = (v: number) => ({ numeric: v, text: getUnitValue(v / 1000000000, NANOSECOND_UNITS, 'seconds') });
    }

    return display;
  }, [profileTypeId]);

  const createFrameFromTable = useCallback(() => {
    const topTable = sortLevelsIntoTable();
    const display = getDisplay();

    const labelValues = new ArrayVector(Object.keys(topTable));
    const selfValues = new ArrayVector(
      Object.values(topTable).map((x) => {
        return parseInt(x.self, 10);
      })
    );
    const totalValues = new ArrayVector(
      Object.values(topTable).map((x) => {
        return parseInt(x.value, 10);
      })
    );

    let df: DataFrame = { fields: [], length: labelValues.length };
    df.fields.push({
      values: labelValues,
      name: 'Symbol',
      type: FieldType.string,
      config: {},
      display: getRawDisplayProcessor(),
    });
    df.fields.push({
      values: selfValues,
      name: 'Self',
      type: FieldType.number,
      config: {
        custom: {
          width: 120,
        },
      },
      display: display,
    });
    df.fields.push({
      values: totalValues,
      name: 'Total',
      type: FieldType.number,
      config: {
        custom: {
          width: 120,
        },
      },
      display: display,
    });
    setDf(df);
  }, [getDisplay, sortLevelsIntoTable]);

  useEffect(() => {
    createFrameFromTable();
  }, [sortLevelsIntoTable, levels, profileTypeId, createFrameFromTable]);

  return (
    <>
      {df.fields && (
        <div className={styles.topTable}>
          <AutoSizer style={{ width: '100%', height: PIXELS_PER_LEVEL * levels.length + 'px' }}>
            {({ width, height }) => (
              <Table width={width} height={height} data={df} initialSortBy={[{ displayName: 'Self', desc: true }]} />
            )}
          </AutoSizer>
        </div>
      )}
    </>
  );
};

const getStyles = (selectedView: SelectedView) => ({
  topTable: css`
    cursor: pointer;
    float: left;
    margin-right: 20px;
    width: ${selectedView === SelectedView.TopTable ? '100%' : 'calc(50% - 20px)'};
  `,
});

export default FlameGraphTopTable;
