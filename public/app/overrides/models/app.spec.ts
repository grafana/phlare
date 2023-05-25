import { brandQuery } from '@webapp/models/query';
import { App, appToQuery } from './app';

it('converts an App to a query', () => {
  expect(
    appToQuery({
      __profile_type__: 'memory:alloc_space:bytes::',
      pyroscope_app: 'simple.golang.app',
      query: '',
      __type__: '',
      name: '',
    })
  ).toEqual(
    brandQuery('memory:alloc_space:bytes::{pyroscope_app="simple.golang.app"}')
  );
});
