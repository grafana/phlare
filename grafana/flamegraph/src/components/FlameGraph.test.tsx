import { screen } from '@testing-library/dom';
import { render } from '@testing-library/react';
import React, { useState } from 'react';

import FlameGraph from './FlameGraph';
import { data } from '../data';
import { MutableDataFrame } from '@grafana/data';
import 'jest-canvas-mock';

describe('FlameGraph', () => {
  const FlameGraphWithProps = () => {
    const [topLevelIndex, setTopLevelIndex] = useState(0);
    const [rangeMin, setRangeMin] = useState(0);
    const [rangeMax, setRangeMax] = useState(1);
    const [query, setQuery] = useState('');

    const flameGraphData = new MutableDataFrame({
      name: 'flamegraph',
      fields: [{ name: 'levels', values: data.flamebearer.levels.map((l) => JSON.stringify(l)) }],
    });
    flameGraphData.meta = {
      custom: {
        Names: data.flamebearer.names,
        Total: data.flamebearer.numTicks,
      },
    };

    return (
      <FlameGraph
        data={flameGraphData}
        topLevelIndex={topLevelIndex}
        rangeMin={rangeMin}
        rangeMax={rangeMax}
        query={query}
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
        setQuery={setQuery}
      />
    );
  };

  it('should render without error', async () => {
    expect(() => render(<FlameGraphWithProps />)).not.toThrow();
  });

  it('should render correctly', async () => {
    Object.defineProperty(HTMLCanvasElement.prototype, 'clientWidth', { value: 1600 });
    render(<FlameGraphWithProps />);
    
    const canvas = screen.getByTestId("flamegraph") as HTMLCanvasElement;
    const ctx = canvas!.getContext('2d');
    const calls = ctx!.__getDrawCalls();
    expect(calls).toMatchSnapshot();
  });
});
