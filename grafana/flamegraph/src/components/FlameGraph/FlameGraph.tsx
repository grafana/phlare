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

import { DataFrame } from '@grafana/data';

import { COLLAPSE_THRESHOLD, PIXELS_PER_LEVEL, MIN_WIDTH_TO_SHOW_TOP_TABLE } from '../../constants';
import { getBarX, getRectDimensionsForLevel, renderRect } from './rendering';
import { ItemWithStart } from './dataTransform';
import FlameGraphTooltip, { getTooltipData } from './FlameGraphTooltip';
import { TooltipData, SelectedView } from '../types';

type Props = {
  data: DataFrame;
  levels: ItemWithStart[][];
  topLevelIndex: number;
  rangeMin: number;
  rangeMax: number;
  query: string;
  setTopLevelIndex: (level: number) => void;
  setRangeMin: (range: number) => void;
  setRangeMax: (range: number) => void;
  selectedView: SelectedView;
  setSelectedView: (view: SelectedView) => void;
  windowWidth: number;
};

const FlameGraph = ({
  data,
  levels,
  topLevelIndex,
  rangeMin,
  rangeMax,
  query,
  setTopLevelIndex,
  setRangeMin,
  setRangeMax,
  selectedView,
  setSelectedView,
  windowWidth,
}: Props) => {
  const styles = getStyles(selectedView, windowWidth);
  const totalTicks = data.fields[1].values.get(0);
  const profileTypeId = data.meta!.custom!.ProfileTypeID;

  const graphRef = useRef<HTMLCanvasElement>(null);
  const tooltipRef = useRef<HTMLDivElement>(null);
  const [tooltipData, setTooltipData] = useState<TooltipData>();
  const [showTooltip, setShowTooltip] = useState(false);

  // Convert pixel coordinates to bar coordinates in the levels array so that we can add mouse events like clicks to
  // the canvas.
  const convertPixelCoordinatesToBarCoordinates = useCallback(
    (x: number, y: number, pixelsPerTick: number) => {
      const levelIndex = Math.floor(y / PIXELS_PER_LEVEL);
      const barIndex = getBarIndex(x, levels[levelIndex], pixelsPerTick, totalTicks, rangeMin);
      return { levelIndex, barIndex };
    },
    [levels, totalTicks, rangeMin]
  );

  const render = useCallback(
    (pixelsPerTick: number) => {
      if (!levels.length) {
        return;
      }
      const ctx = graphRef.current?.getContext('2d')!;
      const graph = graphRef.current!;

      graph.height = PIXELS_PER_LEVEL * levels.length;
      graph.width = graph.clientWidth;
      ctx.textBaseline = 'middle';
      ctx.font = '13.5px Roboto Mono, monospace';
      ctx.strokeStyle = 'white';

      for (let levelIndex = 0; levelIndex < levels.length; levelIndex++) {
        const level = levels[levelIndex];
        // Get all the dimensions of the rectangles for the level. We do this by level instead of per rectangle, because
        // sometimes we collapse multiple bars into single rect.
        const dimensions = getRectDimensionsForLevel(level, levelIndex, totalTicks, rangeMin, pixelsPerTick);
        for (const rect of dimensions) {
          // Render each rectangle based on the computed dimensions
          renderRect(ctx, rect, totalTicks, rangeMin, rangeMax, query, levelIndex, topLevelIndex);
        }
      }
    },
    [levels, query, rangeMax, rangeMin, topLevelIndex, totalTicks]
  );

  useEffect(() => {
    if (graphRef.current) {
      const pixelsPerTick = graphRef.current.clientWidth / totalTicks / (rangeMax - rangeMin);
      render(pixelsPerTick);

      // Clicking allows user to "zoom" into the flamegraph. Zooming means the x axis gets smaller so that the clicked
      // bar takes 100% of the x axis.
      graphRef.current.onclick = (e) => {
        const pixelsPerTick = graphRef.current!.clientWidth / totalTicks / (rangeMax - rangeMin);
        const { levelIndex, barIndex } = convertPixelCoordinatesToBarCoordinates(e.offsetX, e.offsetY, pixelsPerTick);
        
        if (barIndex !== -1 && !isNaN(levelIndex) && !isNaN(barIndex)) {
          setTopLevelIndex(levelIndex);
          setRangeMin(levels[levelIndex][barIndex].start / totalTicks);
          setRangeMax((levels[levelIndex][barIndex].start + levels[levelIndex][barIndex].value) / totalTicks);
        }
      };

      graphRef.current!.onmousemove = (e) => {
        if (tooltipRef.current) {
          setShowTooltip(false);
          const pixelsPerTick = graphRef.current!.clientWidth / totalTicks / (rangeMax - rangeMin);
          const { levelIndex, barIndex } = convertPixelCoordinatesToBarCoordinates(e.offsetX, e.offsetY, pixelsPerTick);
        
          if (barIndex !== -1 && !isNaN(levelIndex) && !isNaN(barIndex)) {
            tooltipRef.current.style.left = e.clientX + 10 + 'px';
            tooltipRef.current.style.top = e.clientY + 40 + 'px';

            const bar = levels[levelIndex][barIndex];
            const tooltipData = getTooltipData(profileTypeId, bar.label, bar.value, totalTicks);
            setTooltipData(tooltipData);
            setShowTooltip(true);
          }
        }
      };

      graphRef.current!.onmouseleave = () => {
        setShowTooltip(false);
      };
    }
  }, [
    render,
    convertPixelCoordinatesToBarCoordinates,
    levels,
    rangeMin,
    rangeMax,
    topLevelIndex,
    totalTicks,
    windowWidth,
    setTopLevelIndex,
    setRangeMin,
    setRangeMax,
    profileTypeId,
    selectedView,
  ]);

  // If user resizes window with top table as the selected view
  useEffect(() => {
    if (windowWidth < MIN_WIDTH_TO_SHOW_TOP_TABLE && selectedView === SelectedView.TopTable) {
      setSelectedView(SelectedView.FlameGraph);
    }
  }, [selectedView, setSelectedView, windowWidth]);

  return (
    <>
      {selectedView !== SelectedView.TopTable && (
        <canvas className={styles.graph} ref={graphRef} data-testid="flamegraph" />
      )}
      <FlameGraphTooltip tooltipRef={tooltipRef} tooltipData={tooltipData!} showTooltip={showTooltip} />
    </>
  );
};

const getStyles = (selectedView: SelectedView, windowWidth: number) => ({
  graph: css`
    cursor: pointer;
    float: left;
    width: ${selectedView !== SelectedView.FlameGraph && windowWidth >= MIN_WIDTH_TO_SHOW_TOP_TABLE ? '50%' : '100%'};
  `,
});

/**
 * Binary search for a bar in a level, based on the X pixel coordinate. Useful for detecting which bar did user click
 * on.
 */
const getBarIndex = (
  x: number,
  level: ItemWithStart[],
  pixelsPerTick: number,
  totalTicks: number,
  rangeMin: number
) => {
  if (level) {
    let start = 0;
    let end = level.length - 1;

    while (start <= end) {
      const midIndex = (start + end) >> 1;
      const startOfBar = getBarX(level[midIndex].start, totalTicks, rangeMin, pixelsPerTick);
      const startOfNextBar = getBarX(
        level[midIndex].start + level[midIndex].value,
        totalTicks,
        rangeMin,
        pixelsPerTick
      );

      if (startOfBar <= x && startOfNextBar >= x) {
        return startOfNextBar - startOfBar > COLLAPSE_THRESHOLD ? midIndex : -1;
      }

      if (startOfBar > x) {
        end = midIndex - 1;
      } else {
        start = midIndex + 1;
      }
    }
  }
  return -1;
};

export default FlameGraph;
