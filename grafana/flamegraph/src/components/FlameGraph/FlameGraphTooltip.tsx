import { css } from '@emotion/css';
import React, { LegacyRef } from 'react';

import { useStyles, Tooltip } from '@grafana/ui';

import { BYTE_UNITS, COUNT_UNITS, NANOSECOND_UNITS, TooltipData, SampleUnit } from '../types';

type Props = {
  tooltipRef: LegacyRef<HTMLDivElement>;
  tooltipData: TooltipData;
  showTooltip: boolean;
};

const FlameGraphTooltip = ({ tooltipRef, tooltipData, showTooltip }: Props) => {
  const styles = useStyles(getStyles);

  return (
    <div ref={tooltipRef} className={styles.tooltip}>
      {tooltipData && (
        <Tooltip
          content={
            <div>
              <div className={styles.name}>{tooltipData.name}</div>
              <div>
                {tooltipData.percentTitle}: <b>{tooltipData.percentValue}%</b>
              </div>
              <div>
                {tooltipData.unitTitle}: <b>{tooltipData.unitValue}</b>
              </div>
              <div>
                Samples: <b>{tooltipData.samples}</b>
              </div>
            </div>
          }
          placement={'right'}
          show={showTooltip}
        >
          <span></span>
        </Tooltip>
      )}
    </div>
  );
};

export const getTooltipData = (
  profileTypeId: string,
  label: string,
  value: number,
  totalTicks: number
): TooltipData => {
  let samples = value;
  let percentTitle = '';
  let unitTitle = '';
  let unitValue = '';

  const sampleUnit = profileTypeId?.split(':').length === 5 ? profileTypeId.split(':')[2] : '';
  const percent = Math.round(10000 * (samples / totalTicks)) / 100;

  switch (sampleUnit) {
    case SampleUnit.Bytes:
      unitValue = getUnitValue(samples, BYTE_UNITS);
      percentTitle = '% of total RAM';
      unitTitle = 'RAM';
      break;

    case SampleUnit.Count:
      unitValue = getUnitValue(samples, COUNT_UNITS);
      percentTitle = '% of total objects';
      unitTitle = 'Allocated objects';
      break;

    case SampleUnit.Nanoseconds:
      unitValue = getUnitValue(
        // convert nanoseconds to seconds
        samples / 1000000000,
        NANOSECOND_UNITS,
        'seconds'
      );
      percentTitle = '% of total time';
      unitTitle = 'Time';
  }

  return {
    name: label,
    percentTitle: percentTitle,
    percentValue: percent,
    unitTitle: unitTitle,
    unitValue: unitValue,
    samples: samples.toLocaleString(),
  };
};

export const getUnitValue = (samples: number, units: any, fallbackSuffix = '') => {
  let unitValue: string;
  let suffix = '';

  for (let unit of units) {
    if (samples >= unit.divider) {
      suffix = unit.suffix;
      samples = samples / unit.divider;
    } else {
      break;
    }
  }

  unitValue = samples.toString();
  if (unitValue.toString().includes('.')) {
    const afterDot = unitValue.toString().split('.')[1];
    if (afterDot.length > 2) {
      unitValue = samples.toFixed(2);
    }
  }

  unitValue += ' ' + (suffix !== '' ? suffix : fallbackSuffix);

  return unitValue;
};

const getStyles = () => ({
  tooltip: css`
    position: fixed;
  `,
  name: css`
    margin-bottom: 10px;
  `,
});

export default FlameGraphTooltip;
