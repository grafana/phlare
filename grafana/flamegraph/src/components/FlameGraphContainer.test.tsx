import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React from 'react';

import { data } from './FlameGraph/testData/dataNestedSet';
import { MutableDataFrame } from '@grafana/data';
import FlameGraphContainer from './FlameGraphContainer';

describe('FlameGraphContainer', () => {
  const FlameGraphContainerWithProps = () => {
    const flameGraphData = new MutableDataFrame(data);
    flameGraphData.meta = {
      custom: {
        ProfileTypeID: 'cpu:foo:bar',
      },
    };

    return (
      <FlameGraphContainer
        data={flameGraphData}
      />
    );
  };

  it('should render without error', async () => {
    expect(() => render(<FlameGraphContainerWithProps />)).not.toThrow();
  });

  it('should update search when row selected in top table', async () => {
    Object.defineProperty(HTMLCanvasElement.prototype, 'clientWidth', { value: 1600 });
    // Needed for AutoSizer to work in test
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, value: 500 });
    Object.defineProperty(HTMLElement.prototype, 'offsetWidth', { configurable: true, value: 500 });

    render(<FlameGraphContainerWithProps />);
    await userEvent.click(screen.getAllByRole('row')[1]);
    expect(screen.getByDisplayValue('net/http.HandlerFunc.ServeHTTP')).toBeInTheDocument();
    await userEvent.click(screen.getAllByRole('row')[2]);
    expect(screen.getByDisplayValue('total')).toBeInTheDocument();
    await userEvent.click(screen.getAllByRole('row')[2]);
    expect(screen.queryByDisplayValue('total')).not.toBeInTheDocument();
  });
});
