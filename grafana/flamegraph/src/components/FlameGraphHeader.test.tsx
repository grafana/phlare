import '@testing-library/jest-dom';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import React, { useState } from 'react';

import FlameGraphHeader from './FlameGraphHeader';
import { SelectedView } from './types';

describe('FlameGraphHeader', () => {
  const FlameGraphHeaderWithProps = () => {
    const [query, setQuery] = useState('');
    const [selectedView, setSelectedView] = useState(SelectedView.Both);

    return (
      <FlameGraphHeader
        query={query}
        setQuery={setQuery}
        setTopLevelIndex={jest.fn()}
        setRangeMin={jest.fn()}
        setRangeMax={jest.fn()}
        selectedView={selectedView}
        setSelectedView={setSelectedView}
        containerWidth={1600}
      />
    );
  };

  it('reset button should remove query text', async () => {
    render(<FlameGraphHeaderWithProps />);
    await userEvent.type(screen.getByPlaceholderText('Search..'), 'abc');
    expect(screen.getByDisplayValue('abc')).toBeInTheDocument();
    screen.getByRole('button', { name: /Reset/i }).click();
    expect(screen.queryByDisplayValue('abc')).not.toBeInTheDocument();
  });
});
