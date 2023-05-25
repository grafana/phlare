import { brandQuery } from '@webapp/models/query';
import { appToQuery, appFromQuery } from './app';

it('converts an App to a query', () => {
  expect(
    appToQuery({
      __profile_type__: 'memory:alloc_space:bytes::',
      pyroscope_app: 'simple.golang.app',
    })
  ).toEqual(
    brandQuery('memory:alloc_space:bytes::{pyroscope_app="simple.golang.app"}')
  );
});

it('converts a query to an App', () => {
  expect(
    appFromQuery(
      brandQuery(
        'memory:alloc_space:bytes::{pyroscope_app="simple.golang.app"}'
      )
    )
  ).toEqual({
    __profile_type__: 'memory:alloc_space:bytes::',
    pyroscope_app: 'simple.golang.app',
  });
});
