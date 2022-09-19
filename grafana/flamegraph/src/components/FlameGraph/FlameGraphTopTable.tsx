import { css } from '@emotion/css';
import React, { useEffect, useState } from 'react';
import AutoSizer from 'react-virtualized-auto-sizer';

import { useStyles2, Table } from '@grafana/ui';

import { PIXELS_PER_LEVEL } from '../../constants';
import { ItemWithStart } from './dataTransform';
import { getUnitValue } from './FlameGraphTooltip';
import { ArrayVector, DataFrame, DisplayProcessor, FieldType, getRawDisplayProcessor } from '@grafana/data';

type Props = {
  levels: ItemWithStart[][];
}

const FlameGraphTopTable = ({ levels }: Props) => {
  const styles = useStyles2(getStyles);
  const [df, setDf] = useState<any>({fields: []})


  useEffect(() => {
    let label, self, value;
    let topTable: { [key: string]: any; } = [];
    let itemWithStart: any;
    
    for (let i = 0; i < levels.length; i++) {
      for (var j = 0; j < Object.values(levels[i]).length; j++) {
        itemWithStart = Object.values(levels[i])[j];
        label = itemWithStart.label;
        self = itemWithStart.self;
        value = itemWithStart.value;
        topTable[label] = topTable[label] || {};
        topTable[label].self = topTable[label].self ? topTable[label].self + self : self;
        topTable[label].value = topTable[label].value ? topTable[label].value + value : value;
      }
    }

    const unitValues = [
      { divider: 1000, suffix: 'K' },
      { divider: 1000, suffix: 'M' },
      { divider: 1000, suffix: 'G' },
      { divider: 1000, suffix: 'T' },
    ];
    const labelValues = new ArrayVector(Object.keys(topTable));
    const selfValues = new ArrayVector(Object.values(topTable).map(x => { return parseInt(x.self, 10) }));
    const totalValues = new ArrayVector(Object.values(topTable).map(x => { return parseInt(x.value, 10) }));
    const display: DisplayProcessor = (v) => ({ numeric: v, text: getUnitValue(parseInt(v, 10), unitValues) });

    let df: DataFrame = { fields: [], length: labelValues.length };
    df.fields.push({
      values: labelValues,
      name: 'Symbol',
      type: FieldType.string,
      config: {},
      display: getRawDisplayProcessor()
    });
    df.fields.push({
      values: selfValues,
      name: 'Self',
      type: FieldType.number,
      config: {
        custom: {
          width: 80
        }
      },
      display: display
    });
    df.fields.push({
      values: totalValues,
      name: 'Total',
      type: FieldType.number,
      config: {
        custom: {
          width: 80
        }
      },
      display: display
    });
    setDf(df);
  }, [levels]);

  return (
    <>
      {df.fields &&
        <div className={styles.topTable}>
           <AutoSizer style={{ width: '100%', height: PIXELS_PER_LEVEL * levels.length + 'px' }}>
            {({ width, height }) => (
              <Table width={width} height={height} data={df} initialSortBy={[{displayName: 'Self', desc: true}]} />
            )}
          </AutoSizer>
        </div>
      }
    </>
  );
}

const getStyles = () => ({
  topTable: css`
    float: left;
    width: 50%;
  `,
});

export default FlameGraphTopTable;
