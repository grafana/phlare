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
              <div className={styles.name}>{tooltipData.name}</div>
              <div>{tooltipData.percentTitle}: <b>{tooltipData.percentValue}%</b></div>
              <div>{tooltipData.unitTitle}: <b>{tooltipData.unitValue}</b></div>
              <div>Samples: <b>{tooltipData.samples}</b></div>
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
    position: fixed;
    z-index: -10;
  `,
  name: css`
    margin-bottom: 10px;
  `,
});

export default FlameGraphTooltip;
