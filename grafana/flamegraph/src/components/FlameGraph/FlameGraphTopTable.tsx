import { css } from '@emotion/css';
import React, { useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import { useStyles2, Table } from '@grafana/ui';

import { PIXELS_PER_LEVEL } from '../../constants';
import { Item } from './dataTransform';
import { getUnitValue } from './FlameGraphTooltip';
import { BYTE_UNITS, COUNT_UNITS, NANOSECOND_UNITS, SampleUnit } from '../types';
import { ArrayVector, DataFrame, DisplayProcessor, FieldType, getRawDisplayProcessor } from '@grafana/data';

type Props = {
  levels: Item[][];
  profileTypeId: string;
};

const FlameGraphTopTable = ({ levels, profileTypeId }: Props) => {
  const styles = useStyles2(getStyles);
  const [df, setDf] = useState<DataFrame>({ fields: [], length: 0 });

  useEffect(() => {
    let label, self, value;
    let unitValues: Array<{ divider: number; suffix: string; }>;
    let topTable: { [key: string]: any } = [];
    let item: Item;

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

    const sampleUnit = profileTypeId?.split(':').length === 5 ? profileTypeId.split(':')[2] : '';
    switch (sampleUnit) {
      case SampleUnit.Bytes:
        unitValues = BYTE_UNITS;
        break;

      case SampleUnit.Count:
        unitValues = COUNT_UNITS;
        break;

      case SampleUnit.Nanoseconds:
        unitValues = NANOSECOND_UNITS;
    }
    const display: DisplayProcessor = (v) => ({ numeric: v, text: getUnitValue(parseInt(v, 10), unitValues) });

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
          width: 80,
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
          width: 80,
        },
      },
      display: display,
    });
    setDf(df);
  }, [levels, profileTypeId]);

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

const getStyles = () => ({
  topTable: css`
    float: left;
    width: 50%;
  `,
});

export default FlameGraphTopTable;
