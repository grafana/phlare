import { screen } from '@testing-library/dom';
import { render } from '@testing-library/react';
import React, { useState } from 'react';

import FlameGraph from './FlameGraph';
import { SelectedView } from '../types';
import { data } from './testData/dataNestedSet';
import { DataFrameView, MutableDataFrame } from '@grafana/data';
import 'jest-canvas-mock';
import { Item, nestedSetToLevels } from './dataTransform';

jest.mock('react-use', () => ({
  useMeasure: () => {
    const ref = React.useRef();
    return [ref, { width: 1600 }];
  },
}));

describe('FlameGraph', () => {
  const FlameGraphWithProps = () => {
    const [topLevelIndex, setTopLevelIndex] = useState(0);
    const [rangeMin, setRangeMin] = useState(0);
    const [rangeMax, setRangeMax] = useState(1);
    const [query] = useState('');
    const [selectedView, _] = useState(SelectedView.Both);

    const flameGraphData = new MutableDataFrame(data);
    const dataView = new DataFrameView<Item>(flameGraphData);
    const levels = nestedSetToLevels(dataView);

    return (
      <FlameGraph
        data={flameGraphData}
        levels={levels}
        topLevelIndex={topLevelIndex}
        rangeMin={rangeMin}
        rangeMax={rangeMax}
        query={query}
        setTopLevelIndex={setTopLevelIndex}
        setRangeMin={setRangeMin}
        setRangeMax={setRangeMax}
        selectedView={selectedView}
        windowWidth={1600}
      />
    );
  };

  it('should render without error', async () => {
    expect(() => render(<FlameGraphWithProps />)).not.toThrow();
  });

  it('should render correctly', async () => {
    render(<FlameGraphWithProps />);

    const canvas = screen.getByTestId('flameGraph') as HTMLCanvasElement;
    const ctx = canvas!.getContext('2d');
    const calls = ctx!.__getDrawCalls();
    expect(calls).toMatchSnapshot();
  });
});
