import { css } from '@emotion/css';
import React, { LegacyRef } from 'react';

import { useStyles, Tooltip } from '@grafana/ui';

import { TooltipData, SampleUnit } from '../types';
import { createTheme, Field, getDisplayProcessor } from '@grafana/data';

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

export const getTooltipData = (field: Field, label: string, value: number, totalTicks: number): TooltipData => {
  let samples = value;
  let percentTitle = '';
  let unitTitle = '';

  const processor = getDisplayProcessor({ field, theme: createTheme() /* theme does not matter for us here */ });
  const displayValue = processor(value);
  const percent = Math.round(10000 * (samples / totalTicks)) / 100;
  let unitValue = displayValue.text + displayValue.suffix;

  switch (field.config.unit) {
    case SampleUnit.Bytes:
      percentTitle = '% of total';
      unitTitle = 'RAM';
      break;
    case SampleUnit.Nanoseconds:
      percentTitle = '% of total time';
      unitTitle = 'Time';
      break;
    default:
      percentTitle = '% of total';
      unitTitle = 'Count';
      if (!displayValue.suffix) {
        // Makes sure we don't show 123undefined or something like that if suffix isn't defined
        unitValue = displayValue.text;
      }
      break;
  }

  return {
    name: label,
    percentTitle: percentTitle,
    percentValue: percent,
    unitTitle: unitTitle,
    unitValue,
    samples: samples.toLocaleString(),
  };
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
