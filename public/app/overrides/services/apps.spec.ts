import { Result } from '@webapp/util/fp';
import { fetchApps } from './apps';
import * as base from '@webapp/services/base';

jest.mock('@webapp/services/base', () => {
  return {
    __esModule: true,
    ...jest.requireActual('@webapp/services/base'),
  };
});

const mockData = {
  labelsSet: [
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:alloc_objects:count::',
        },
        {
          name: '__type__',
          value: 'alloc_objects',
        },
        {
          name: '__unit__',
          value: 'count',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:alloc_objects:count::',
        },
        {
          name: '__type__',
          value: 'alloc_objects',
        },
        {
          name: '__unit__',
          value: 'count',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:alloc_space:bytes::',
        },
        {
          name: '__type__',
          value: 'alloc_space',
        },
        {
          name: '__unit__',
          value: 'bytes',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:alloc_space:bytes::',
        },
        {
          name: '__type__',
          value: 'alloc_space',
        },
        {
          name: '__unit__',
          value: 'bytes',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:inuse_objects:count::',
        },
        {
          name: '__type__',
          value: 'inuse_objects',
        },
        {
          name: '__unit__',
          value: 'count',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:inuse_objects:count::',
        },
        {
          name: '__type__',
          value: 'inuse_objects',
        },
        {
          name: '__unit__',
          value: 'count',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:inuse_space:bytes::',
        },
        {
          name: '__type__',
          value: 'inuse_space',
        },
        {
          name: '__unit__',
          value: 'bytes',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'memory',
        },
        {
          name: '__profile_type__',
          value: 'memory:inuse_space:bytes::',
        },
        {
          name: '__type__',
          value: 'inuse_space',
        },
        {
          name: '__unit__',
          value: 'bytes',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'foo',
          value: 'bar',
        },
        {
          name: 'function',
          value: 'fast',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'foo',
          value: 'bar',
        },
        {
          name: 'function',
          value: 'fast',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'foo',
          value: 'bar',
        },
        {
          name: 'function',
          value: 'slow',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'foo',
          value: 'bar',
        },
        {
          name: 'function',
          value: 'slow',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'foo',
          value: 'bar',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'foo',
          value: 'bar',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
    {
      labels: [
        {
          name: '__delta__',
          value: 'false',
        },
        {
          name: '__name__',
          value: 'process_cpu',
        },
        {
          name: '__period_type__',
          value: 'cpu',
        },
        {
          name: '__period_unit__',
          value: 'nanoseconds',
        },
        {
          name: '__profile_type__',
          value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        },
        {
          name: '__type__',
          value: 'cpu',
        },
        {
          name: '__unit__',
          value: 'nanoseconds',
        },
        {
          name: 'pyroscope_app',
          value: 'simple.golang.app2',
        },
        {
          name: 'pyroscope_spy',
          value: 'gospy',
        },
      ],
    },
  ],
};

//it('smoke', () => {
//  expect(groupByAppAndProfileId(mockData)).toBe(true);
//});
//
describe('appsService', () => {
  afterEach(() => {
    jest.restoreAllMocks();
  });

  describe('error cases', () => {
    test.each([
      [{}, []],
      [null, []],
      [{ labelSets: [] }, []],
    ])(
      `server returned='%s', should transform into %s`,
      async (response, expected) => {
        const spy = jest.spyOn(base, 'requestWithOrgID');
        spy.mockReturnValue(Promise.resolve(Result.ok(response)));

        const res = await fetchApps();
        expect(res.isOk).toBe(true);
        expect(res.value).toMatchObject(expected);
      }
    );
  });

  it('works', async () => {
    const spy = jest.spyOn(base, 'requestWithOrgID');
    spy.mockReturnValue(Promise.resolve(Result.ok(mockData)));

    const res = await fetchApps();
    expect(res.isOk).toBe(true);
    expect(res.value).toMatchInlineSnapshot(`
      [
        {
          "__profile_type__": "memory:alloc_objects:count::",
          "name": "simple.golang.app",
          "pyroscope_app": "simple.golang.app",
        },
        {
          "__profile_type__": "memory:alloc_objects:count::",
          "name": "simple.golang.app2",
          "pyroscope_app": "simple.golang.app2",
        },
        {
          "__profile_type__": "memory:alloc_space:bytes::",
          "name": "simple.golang.app",
          "pyroscope_app": "simple.golang.app",
        },
        {
          "__profile_type__": "memory:alloc_space:bytes::",
          "name": "simple.golang.app2",
          "pyroscope_app": "simple.golang.app2",
        },
        {
          "__profile_type__": "memory:inuse_objects:count::",
          "name": "simple.golang.app",
          "pyroscope_app": "simple.golang.app",
        },
        {
          "__profile_type__": "memory:inuse_objects:count::",
          "name": "simple.golang.app2",
          "pyroscope_app": "simple.golang.app2",
        },
        {
          "__profile_type__": "memory:inuse_space:bytes::",
          "name": "simple.golang.app",
          "pyroscope_app": "simple.golang.app",
        },
        {
          "__profile_type__": "memory:inuse_space:bytes::",
          "name": "simple.golang.app2",
          "pyroscope_app": "simple.golang.app2",
        },
        {
          "__profile_type__": "process_cpu:cpu:nanoseconds:cpu:nanoseconds",
          "name": "simple.golang.app",
          "pyroscope_app": "simple.golang.app",
        },
        {
          "__profile_type__": "process_cpu:cpu:nanoseconds:cpu:nanoseconds",
          "name": "simple.golang.app2",
          "pyroscope_app": "simple.golang.app2",
        },
      ]
    `);
  });

  // For example, if a cpu profile contains different tags
  // The server will return with that level of granularity
  // Which is not required to build an "App"
  it('remove duplicates from same _profile_type__/name pair', async () => {
    const spy = jest.spyOn(base, 'requestWithOrgID');
    const mockData = {
      labelsSet: [
        {
          labels: [
            {
              name: '__profile_type__',
              value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
            },
            { name: 'pyroscope_app', value: 'simple.golang.app' },
          ],
        },
        {
          labels: [
            {
              name: '__profile_type__',
              value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
            },
            { name: 'pyroscope_app', value: 'simple.golang.app2' },
          ],
        },
        {
          labels: [
            {
              name: '__profile_type__',
              value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
            },
            { name: 'pyroscope_app', value: 'simple.golang.app' },
            { name: 'function', value: 'fast' },
          ],
        },
        {
          labels: [
            {
              name: '__profile_type__',
              value: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
            },
            { name: 'pyroscope_app', value: 'simple.golang.app' },
            { name: 'function', value: 'slow' },
          ],
        },
      ],
    };
    spy.mockReturnValue(Promise.resolve(Result.ok(mockData)));

    const res = await fetchApps();
    expect(res.isOk).toBe(true);
    expect(res.value).toMatchObject([
      {
        __profile_type__: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        pyroscope_app: 'simple.golang.app',
      },
      {
        __profile_type__: 'process_cpu:cpu:nanoseconds:cpu:nanoseconds',
        pyroscope_app: 'simple.golang.app2',
      },
    ]);
  });
});
