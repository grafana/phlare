// This component is based on logic from the flamebearer project
// https://github.com/mapbox/flamebearer

// ISC License

// Copyright (c) 2018, Mapbox

// Permission to use, copy, modify, and/or distribute this software for any purpose
// with or without fee is hereby granted, provided that the above copyright notice
// and this permission notice appear in all copies.

// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
// REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
// FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
// INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
// OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
// TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
// THIS SOFTWARE.
import { css } from '@emotion/css';
import React, { CSSProperties, useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { useWindowSize } from 'react-use';

import { DataFrame } from '@grafana/data';
import { colors, fuzzyMatch, useStyles2 } from '@grafana/ui';

import {
  BAR_BORDER_WIDTH,
  COLLAPSE_THRESHOLD,
  HIDE_THRESHOLD,
  LABEL_THRESHOLD,
  NAME_OFFSET,
  PIXELS_PER_LEVEL,
  STEP_OFFSET,
} from '../constants';

type Props = {
  data: DataFrame;
  topLevelIndex: number;
  rangeMin: number;
  rangeMax: number;
  query: string;
  setTopLevelIndex: (level: number) => void;
  setRangeMin: (range: number) => void;
  setRangeMax: (range: number) => void;
};

const FlameGraph = ({
  data,
  topLevelIndex,
  rangeMin,
  rangeMax,
  query,
  setTopLevelIndex,
  setRangeMin,
  setRangeMax,
}: Props) => {
  const styles = useStyles2(getStyles);

  const levels = useLevels(data);
  const names = data.meta!.custom!.Names;
  const totalTicks = data.meta!.custom!.Total;

  const { width: windowWidth } = useWindowSize();
  const graphRef = useRef<HTMLDivElement>(null);
  const [bars, setBars] = useState<any>([]);

  // get the x coordinate of the bar i.e. where it starts on the vertical plane
  const getBarX = useCallback(
    (accumulatedTicks: number, pixelsPerTick: number) => {
      // rangeTicks = totalTicks * rangeMin;
      return (accumulatedTicks - totalTicks * rangeMin) * pixelsPerTick;
    },
    [rangeMin, totalTicks]
  );

  const getBarColor = (h: number, l: number) => {
    return `hsl(${h}, 100%, ${l}%)`;
  };

  useEffect(() => {
    if (levels) {
      for (const level of levels) {
        let prev = 0;
        for (let i = 0; i < level.length; i += STEP_OFFSET) {
          level[i] += prev;
          prev = level[i] + level[i + 1];
        }
      }
    }
  }, [levels]);

  const render = useCallback(
    (pixelsPerTick: number) => {
      if (!levels) {
        return;
      }
      let bars = [];
      let level, barX, curBarTicks, collapsed, width, name, queryResult, intensity, h, l;
      let style: CSSProperties;

      const graph = graphRef.current!;
      graph.style.height = PIXELS_PER_LEVEL * levels.length + 'px';

      for (let levelIndex = 0; levelIndex < levels.length; levelIndex++) {
        level = levels[levelIndex];

        for (let barIndex = 0; barIndex < level.length; barIndex += STEP_OFFSET) {
          // accumulatedBarTicks = level[barIndex];
          barX = getBarX(level[barIndex], pixelsPerTick);
          if (barX + BAR_BORDER_WIDTH * 2 > graphRef.current!.clientWidth) {
            continue;
          }
          curBarTicks = level[barIndex + 1];

          // merge very small blocks into big "collapsed" ones for performance
          collapsed = curBarTicks * pixelsPerTick <= COLLAPSE_THRESHOLD;
          if (collapsed) {
            while (
              barIndex < level.length - STEP_OFFSET &&
              level[barIndex] + curBarTicks === level[barIndex + STEP_OFFSET] &&
              level[barIndex + STEP_OFFSET + 1] * pixelsPerTick <= COLLAPSE_THRESHOLD
            ) {
              barIndex += STEP_OFFSET;
              curBarTicks += level[barIndex + 1];
            }
          }

          width = curBarTicks * pixelsPerTick;
          if (barX < 0) {
            width = barX + width;
            barX = 0;
          }
          if (width < HIDE_THRESHOLD) {
            continue;
          }

          style = {
            left: barX,
            top: levelIndex * PIXELS_PER_LEVEL,
            width: width,
          };

          //  / (rangeMax - rangeMin) here so when you click a bar it will adjust the top (clicked)bar to the most 'intense' color
          intensity = Math.min(1, curBarTicks / totalTicks / (rangeMax - rangeMin));
          h = 50 - 50 * intensity;
          l = 65 + 7 * intensity;

          name = names[level[barIndex + NAME_OFFSET]];
          queryResult = query && fuzzyMatch(name.toLowerCase(), query.toLowerCase()).found;

          if (!collapsed) {
            if (query) {
              style['background'] = queryResult ? getBarColor(h, l) : colors[55];
            } else {
              style['background'] = levelIndex > topLevelIndex - 1 ? getBarColor(h, l) : getBarColor(h, l + 15);
            }
            style['outline'] = BAR_BORDER_WIDTH + 'px solid ' + colors[55];
            bars.push(
              <div key={Math.random()} className={styles.bar} data-x={levelIndex} data-y={barIndex} style={style}>
                {width >= LABEL_THRESHOLD ? name : ''}
              </div>
            );
          } else {
            style['background'] = queryResult ? getBarColor(h, l) : colors[55];
            bars.push(<div key={Math.random()} className={styles.bar} style={style}></div>);
          }
        }
      }

      setBars(bars);
    },
    [levels, getBarX, names, query, totalTicks, rangeMax, rangeMin, topLevelIndex, styles.bar]
  );

  useEffect(() => {
    if (graphRef.current) {
      const pixelsPerTick = graphRef.current.clientWidth / totalTicks / (rangeMax - rangeMin);
      render(pixelsPerTick);

      graphRef.current.onclick = (e) => {
        const levelIndex = parseInt((e as any).target?.getAttribute('data-x'), 10);
        const barIndex = parseInt((e as any).target?.getAttribute('data-y'), 10);

        if (!isNaN(levelIndex) && !isNaN(barIndex)) {
          setTopLevelIndex(levelIndex);
          setRangeMin(levels[levelIndex][barIndex] / totalTicks);
          setRangeMax((levels[levelIndex][barIndex] + levels[levelIndex][barIndex + 1]) / totalTicks);
        }
      };
    }
  }, [
    render,
    levels,
    names,
    rangeMin,
    rangeMax,
    topLevelIndex,
    totalTicks,
    windowWidth,
    setTopLevelIndex,
    setRangeMin,
    setRangeMax,
  ]);

  return (
    <div className={styles.graph} ref={graphRef} data-testid="flamegraph">
      {bars}
    </div>
  );
};

function useLevels(frame: DataFrame) {
  return useMemo(() => {
    const levels: number[][] = [];
    const levelsField = frame.fields.find((f) => f.name === 'levels');
    if (!levelsField) {
      return [];
    }
    for (let i = 0; i < levelsField.values.length; i++) {
      levels.push(JSON.parse(levelsField.values.get(i)));
    }
    return levels;
  }, [frame]);
}

const getStyles = () => ({
  graph: css`
    position: relative;
    overflow: hidden;
    font-family: 'Roboto';
    font-size: 13px;
    text-indent: 3px;
    white-space: nowrap;
  `,
  bar: css`
    position: absolute;
    color: #222;
    cursor: pointer;
    height: ${PIXELS_PER_LEVEL}px;
    overflow: hidden;
  `,
});

export default FlameGraph;
