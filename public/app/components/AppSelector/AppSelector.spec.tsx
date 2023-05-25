import React from 'react';
import { AppSelector } from './AppSelector';
import { render, screen } from '@testing-library/react';
import { brandQuery } from '@webapp/models/query';

describe('AppSelector', () => {
  it('works', () => {
    render(
      <AppSelector
        apps={[]}
        onSelected={() => {}}
        selectedQuery={brandQuery('')}
      />
    );

    expect(screen.getByRole('button')).toHaveTextContent('TEST');
  });
});
