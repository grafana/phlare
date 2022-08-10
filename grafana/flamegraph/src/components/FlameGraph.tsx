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
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useWindowSize } from 'react-use';

import { colors, useStyles } from '@grafana/ui';

import  { BAR_BORDER_WIDTH, BAR_TEXT_PADDING_LEFT, COLLAPSE_THRESHOLD, HIDE_THRESHOLD, LABEL_THRESHOLD, NAME_OFFSET, PIXELS_PER_LEVEL, STEP_OFFSET } from '../Constants';
import { data } from '../Data';
import FlameGraphHeader from './FlameGraphHeader';

const FlameGraph = () => {
  const styles = useStyles(getStyles);
  const levels = data['flamebearer']['levels'];
  const names = data['flamebearer']['names'];
  const totalTicks = data['flamebearer']['numTicks'];

  const { width: windowWidth } = useWindowSize();
  const graphRef = useRef<HTMLCanvasElement>(null);
  const [topLevelIndex, setTopLevelIndex] = useState(0)
  const [rangeMin, setRangeMin] = useState(0)
  const [rangeMax, setRangeMax] = useState(1)

  
  // get the x coordinate of the bar i.e. where it starts on the vertical plane
  const getBarX = useCallback((barTicks: number, pixelsPerTick: number) => {
    return (barTicks - totalTicks * rangeMin) * pixelsPerTick;
  }, [rangeMin, totalTicks]);

  // binary search for a bar in a level
  const getBarIndex = useCallback((x: number, level: number[], pixelsPerTick: number) => {
    if (level) {
      let start = 0;
      let end = level.length;

      while (start <= end) {
        const s = start / STEP_OFFSET;
        const e = end / STEP_OFFSET;
        // we divide by STEP_OFFSET above (and multiple by it for our midIndex const) because we don't want to have an uneven mid point
        // after performing our bitwise right shift. STEP_OFFSET === 4 and we are modifying the end/start by taking/adding STEP_OFFSET below
        // i.e. each set of 4 values in the level array is the data needed to render a bar. When used with our STEP_OFFSET (+/- 4), 
        // an uneven midIndex would not always represent one bar i.e. one block of 4 sequential values in our level array.
        // 
        // if we're not moving in blocks of 4 then startOfNextBar could be the same as startOfBar, 
        // i.e. we would never return the index of the bar we're searching for
        // because in a block of 4, the 0th value is the accumulated ticks so far and the 1st value is this bar ticks
        const midIndex = STEP_OFFSET * ((s + e) >> 1);
        const startOfBar = getBarX(level[midIndex], pixelsPerTick);
        const startOfNextBar = getBarX(level[midIndex] + level[midIndex + 1], pixelsPerTick);
        console.log(e, midIndex, startOfBar, startOfNextBar);
        
        if (startOfBar <= x && startOfNextBar >= x) {
          return startOfNextBar - startOfBar > COLLAPSE_THRESHOLD ? midIndex : -1;
        }

        if (startOfBar > x) {
          end = midIndex - STEP_OFFSET;
        } else {
          start = midIndex + STEP_OFFSET;
        }
      }
    }
    return -1;
  }, [getBarX]);

  // convert pixel coordinates to bar coordinates in the levels array
  const convertPixelCoordinatesToBarCoordinates = useCallback((x: number, y: number, pixelsPerTick: number) => {
    const levelIndex = Math.floor(y / PIXELS_PER_LEVEL);
    const barIndex = getBarIndex(x, levels[levelIndex], pixelsPerTick);
    return {levelIndex, barIndex};
  }, [getBarIndex, levels]);

  const getBarColor = (h: number, l: number) => {
    return `hsl(${h}, 100%, ${l}%)`;
  };

  const render = useCallback((pixelsPerTick: number) => {
    if (!levels) {return;}
    const ctx = graphRef.current?.getContext('2d')!;
    const graph = graphRef.current!;
    graph.height = PIXELS_PER_LEVEL * (levels.length);
    graph.width = graph.clientWidth;
    ctx.textBaseline = 'middle';
    ctx.font = '13px Roboto, sans-serif';
    ctx.strokeStyle = 'white';

    for (let levelIndex = 0; levelIndex < levels.length; levelIndex++) {
      const level = levels[levelIndex];

      for (let barIndex = 0; barIndex < level.length; barIndex += STEP_OFFSET) {
        const accumulatedBarTicks = level[barIndex];
        const barX = getBarX(accumulatedBarTicks, pixelsPerTick);
        const name = names[level[barIndex + NAME_OFFSET]];
        let curBarTicks = level[barIndex + 1];

        // merge very small blocks into big "collapsed" ones for performance
        const collapsed = curBarTicks * pixelsPerTick <= COLLAPSE_THRESHOLD;
        if (collapsed) {
          while (
            barIndex < level.length - STEP_OFFSET &&
            accumulatedBarTicks + curBarTicks === level[barIndex + STEP_OFFSET] &&
            level[barIndex + STEP_OFFSET + 1] * pixelsPerTick <= COLLAPSE_THRESHOLD
          ) {
            barIndex += STEP_OFFSET;
            curBarTicks += level[barIndex + 1];
          }
        }

        const width = curBarTicks * pixelsPerTick - (collapsed ? 0 : BAR_BORDER_WIDTH * 2);
        if (width < HIDE_THRESHOLD) {continue;}

        ctx.beginPath();                
        ctx.rect(barX + (collapsed ? 0 : BAR_BORDER_WIDTH), levelIndex * PIXELS_PER_LEVEL, width, PIXELS_PER_LEVEL);

        //  / (rangeMax - rangeMin) here so when you click a bar it will adjust the top (clicked)bar to the most 'intense' color
        const intensity = Math.min(1, (curBarTicks / totalTicks) / (rangeMax - rangeMin));
        const h = 50 - (50 * intensity);
        const l = 65 + (7 * intensity);

        if (!collapsed) {
          ctx.stroke();
          ctx.fillStyle = levelIndex > topLevelIndex - 1 ? getBarColor(h, l) : getBarColor(h, l + 15);
        } else {
          ctx.fillStyle = colors[55];
        }
        ctx.fill();

        if (!collapsed && width >= LABEL_THRESHOLD) {
          ctx.save();
          ctx.clip(); // so text does not overflow
          ctx.fillStyle = 'black';
          ctx.fillText(`${name}`, Math.max(barX, 0) + BAR_TEXT_PADDING_LEFT, levelIndex * PIXELS_PER_LEVEL + PIXELS_PER_LEVEL / 2);
          ctx.restore();
        }
      }
    }
  }, [getBarX, levels, names, rangeMin, rangeMax, topLevelIndex, totalTicks]);

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

  useEffect(() => {
    if (graphRef.current) {
      const pixelsPerTick = graphRef.current.clientWidth / totalTicks / (rangeMax - rangeMin);
      render(pixelsPerTick);

      graphRef.current.onclick = (e) => {
        const pixelsPerTick = graphRef.current!.clientWidth / totalTicks / (rangeMax - rangeMin);
        const {levelIndex, barIndex} = convertPixelCoordinatesToBarCoordinates(e.offsetX, e.offsetY, pixelsPerTick);
        if (barIndex === -1) {return;}
        if (!isNaN(levelIndex) && !isNaN(barIndex)) {
          setTopLevelIndex(levelIndex);
          setRangeMin(levels[levelIndex][barIndex] / totalTicks);
          setRangeMax((levels[levelIndex][barIndex] + levels[levelIndex][barIndex + 1]) / totalTicks);
        }
    };
    }
  }, [convertPixelCoordinatesToBarCoordinates, levels, rangeMin, rangeMax, render, totalTicks, windowWidth]);

  return (
    <>
      <FlameGraphHeader 
        setTopLevelIndex={setTopLevelIndex} 
        setRangeMin={setRangeMin} 
        setRangeMax={setRangeMax}
      />
      <canvas className={styles.graph} ref={graphRef} />
    </>
  );
}

const getStyles = () => ({
  graph: css`
    cursor: pointer;
    width: 100%;
  `,
});

export default FlameGraph;
