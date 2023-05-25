import { groupByAppAndProfileId } from './apps';

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

it('works', () => {
  expect(groupByAppAndProfileId(mockData)).toBe(true);
});
