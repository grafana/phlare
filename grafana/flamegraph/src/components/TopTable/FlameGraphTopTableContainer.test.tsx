import { render, screen } from '@testing-library/react';
import React, { useState } from 'react';

import FlameGraphTopTableContainer from './FlameGraphTopTableContainer';
import { SelectedView } from '../types';
import { data } from '../FlameGraph/testData/dataNestedSet';
import { DataFrameView, MutableDataFrame } from '@grafana/data';
import { Item, nestedSetToLevels } from '../FlameGraph/dataTransform';

describe('FlameGraphTopTableContainer', () => {
  const FlameGraphTopTableContainerWithProps = () => {
    const [search, setSearch] = useState('');
    const [selectedView, _] = useState(SelectedView.Both);

    const flameGraphData = new MutableDataFrame(data);
    const dataView = new DataFrameView<Item>(flameGraphData);
    const levels = nestedSetToLevels(dataView);

    return (
      <FlameGraphTopTableContainer
        data={flameGraphData}
        totalLevels={levels.length}
        selectedView={selectedView}
        search={search}
        setSearch={setSearch}
        setTopLevelIndex={jest.fn()}
        setRangeMin={jest.fn()}
        setRangeMax={jest.fn()}
      />
    );
  };

  it('should render without error', async () => {
    expect(() => render(<FlameGraphTopTableContainerWithProps />)).not.toThrow();
  });

  it('should render correctly', async () => {
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
    expect(cells[1].textContent).toEqual('31.7 K');
    expect(cells[2].textContent).toEqual('31.7 Bil');
    expect(cells[24].textContent).toEqual('github.com/grafana/fire/pkg/fire.(*Fire).initServer.func2.1');
    expect(cells[25].textContent).toEqual('5.58 K');
    expect(cells[26].textContent).toEqual('5.58 Bil');
  });
});
