import { render, screen } from '@testing-library/react';
import React, { useState } from 'react';

import FlameGraphTopTableContainer from './FlameGraphTopTableContainer';
import { SelectedView } from '../types';
import { data } from '../FlameGraph/testData/dataNestedSet';
import { DataFrameView, MutableDataFrame } from '@grafana/data';
import { Item, nestedSetToLevels } from '../FlameGraph/dataTransform';

describe('FlameGraphTopTableContainer', () => {
  const FlameGraphTopTableContainerWithProps = () => {
    const [query, setQuery] = useState('');
    const [selectedView, _] = useState(SelectedView.Both);

    const flameGraphData = new MutableDataFrame(data);
    const dataView = new DataFrameView<Item>(flameGraphData);
    const levels = nestedSetToLevels(dataView);

    return (
      <FlameGraphTopTableContainer
        levels={levels}
        profileTypeId={'memory:alloc_objects:count:space:bytes'}
        selectedView={selectedView}
        query={query}
        setQuery={setQuery}
      />
    );
  };

  it('should render without error', async () => {
    expect(() => render(<FlameGraphTopTableContainerWithProps />)).not.toThrow();
  });

  it('should render correctly', async () => {
    Object.defineProperty(HTMLCanvasElement.prototype, 'clientWidth', { value: 1600 });
    // Needed for AutoSizer to work in test
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, value: 500 });
    Object.defineProperty(HTMLElement.prototype, 'offsetWidth', { configurable: true, value: 500 });

    render(<FlameGraphTopTableContainerWithProps />);
    const rows = screen.getAllByRole('row');
    expect(rows).toHaveLength(17); // + 1 for the columnHeaders

    const columnHeaders = screen.getAllByRole('columnheader');
    expect(columnHeaders).toHaveLength(3);
    expect(columnHeaders[0].textContent).toEqual('Symbol');
    expect(columnHeaders[1].textContent).toEqual('Self');
    expect(columnHeaders[2].textContent).toEqual('Total');

    const cells = screen.getAllByRole('cell');
    expect(cells).toHaveLength(48); // 16 rows
    expect(cells[0].textContent).toEqual('net/http.HandlerFunc.ServeHTTP');
    expect(cells[1].textContent).toEqual('31.69 K');
    expect(cells[2].textContent).toEqual('31.69 G');
    expect(cells[24].textContent).toEqual('github.com/grafana/fire/pkg/fire.(*Fire).initServer.func2.1');
    expect(cells[25].textContent).toEqual('5.58 K');
    expect(cells[26].textContent).toEqual('5.58 G');
  });
});
