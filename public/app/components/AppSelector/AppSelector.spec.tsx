import React from 'react';
import { AppSelector } from './AppSelector';
import { render, screen } from '@testing-library/react';
import { brandQuery } from '@webapp/models/query';
import { appToQuery } from '@phlare/overrides/models/app';

describe('AppSelector', () => {
  describe('when no query exists / is invalid', () => {
    it('renders an empty app selector', () => {
      render(
        <AppSelector
          apps={[]}
          onSelected={() => {}}
          selectedQuery={brandQuery('')}
        />
      );

      expect(screen.getByRole('button')).toHaveTextContent(
        'Select an application'
      );
    });
  });

  describe('when a query exists', () => {
    describe('when an equivalent app exists', () => {
      it('selects that app', () => {
        const apps = [
          {
            __profile_type__: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
            pyroscope_app: 'myapp',
          },
        ];
        const query = appToQuery(apps[0]);

        render(
          <AppSelector
            apps={apps}
            onSelected={() => {}}
            selectedQuery={query}
          />
        );

        expect(screen.getByRole('button')).toHaveTextContent(query);
      });
    });
  });

  describe('when a query exists', () => {
    describe('when an equivalent app DOES NOT exist', () => {
      it('still shows up', () => {
        const apps = [
          {
            __profile_type__: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
            pyroscope_app: 'myapp',
            name: '',
          },
        ];

        const query = brandQuery(
          'memory:alloc_objects:count::1{pyroscope_app="simple.golang.app"}'
        );

        render(
          <AppSelector
            apps={apps}
            onSelected={() => {}}
            selectedQuery={query}
          />
        );

        expect(screen.getByRole('button')).toHaveTextContent(query);
      });
    });
  });

  // TODO: test
  // * filter
  // * interaction
});
