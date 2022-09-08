import { css } from '@emotion/css';
import React, { LegacyRef } from 'react';

import { useStyles, Tooltip } from '@grafana/ui';

import { TooltipData } from './types';

type Props = {
  tooltipRef: LegacyRef<HTMLDivElement>,
	tooltipData: TooltipData,
  showTooltip: boolean,
};

const FlameGraphTooltip = ({tooltipRef, tooltipData, showTooltip}: Props) => {
	const styles = useStyles(getStyles);
	
	return (
		<div ref={tooltipRef} className={styles.tooltip}>
      {tooltipData &&
        <Tooltip 
          content={
            <div>
              <div>{tooltipData.name}</div>
              <div>{tooltipData.percentTitle}: {tooltipData.percentValue}%</div>
              <div>{tooltipData.unitTitle}: {tooltipData.unitValue}</div>
              <div>Samples: {tooltipData.samples}</div>
            </div>
          } 
          placement={'right'} 
          show={showTooltip}
        >
          <span></span>
        </Tooltip>
      }
    </div>
	)
}

const getStyles = () => ({
  tooltip: css`
    label: joey-tooltip;
    position: fixed;
    z-index: -10;
  `,
});

export default FlameGraphTooltip;
