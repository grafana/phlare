import { css } from '@emotion/css';
import React, { LegacyRef } from 'react';

import { useStyles, Tooltip } from '@grafana/ui';

type Props = {
  tooltipRef: LegacyRef<HTMLDivElement>,
	tooltip: string[],
  showTooltip: boolean,
};

const FlameGraphTooltip = ({tooltipRef, tooltip, showTooltip}: Props) => {
	const styles = useStyles(getStyles);
	
	return (
		<div ref={tooltipRef} className={styles.tooltip}>
      <Tooltip 
        content={
          <div>
            <div>{tooltip[0]}</div>
            <div>{tooltip[1]}%</div>
            <div>{tooltip[2]}</div>
            <div>Samples: {tooltip[3]}</div>
          </div>
        } 
        placement={'right'} 
        show={showTooltip}
      >
        <span></span>
      </Tooltip>
    </div>
	)
}

const getStyles = () => ({
  tooltip: css`
    position: fixed;
  `,
});

export default FlameGraphTooltip;
