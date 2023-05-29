import { App, AppSchema } from '@webapp/models/app';
import { Result } from '@webapp/util/fp';
import { z, ZodError } from 'zod';
import type { RequestError } from '@webapp/services/base';
import { parseResponse, requestWithOrgID } from '@webapp/services/base';

type SeriesResponse = {
  labelsSet: [
    {
      labels: Array<{ name: string; value: string }>;
    }
  ];
};

const SeriesResponseSchema = z.preprocess(
  (arg) => {
    const noop = {
      labelsSet: [],
    };

    // The backend may return an empty object ({})
    if (!arg || !('labelsSet' in arg)) {
      return noop;
    }

    return arg;
  },
  z.object({
    labelsSet: z.array(
      z.object({
        labels: z.array(
          z.object({
            name: z.string(),
            value: z.string(),
          })
        ),
      })
    ),
  })
);

z.string()
  .transform((val) => val.length)
  .pipe(z.number().min(5));

const AppsSchema = SeriesResponseSchema.transform((v) => {
  return groupByAppAndProfileId(v);
})
  .pipe(z.array(AppSchema))
  // Remove duplicates
  .transform((v) => {
    // Generate an unique id
    const idFn = (b: (typeof v)[number]) =>
      `${b.__profile_type__}-${b.pyroscope_app}`;

    const visited = new Set<string>();

    return v.filter((b) => {
      if (visited.has(idFn(b))) {
        return false;
      }

      visited.add(idFn(b));
      return true;
    });
  });

// TODO: change after https://github.com/grafana/phlare/pull/710 is merged
const appTag = 'pyroscope_app';

function mergeLabels(v: SeriesResponse['labelsSet'][number]['labels']) {
  return v.reduce((acc, curr) => {
    // TODO: type
    acc[curr.name] = curr.value;
    return acc;
  }, {});
}

function addQueryAndName(v: ReturnType<typeof mergeLabels>) {
  // TODO: type
  const query = `${v['__profile_type__']}{${appTag}="${v[appTag]}"}`;

  return {
    ...v,
    query,
    name: v[appTag],
    //    name: `${v[appTag]}.${v['__profile_type__']}`, // the app selector expects this
  };
}

//export function groupByAppAndProfileId(props: SeriesResponse) {
export function groupByAppAndProfileId(
  props: z.infer<typeof SeriesResponseSchema>
) {
  return props.labelsSet
    .flatMap((v) => mergeLabels(v.labels))
    .map(addQueryAndName);
}

export async function fetchApps(): Promise<
  Result<App[], RequestError | ZodError>
> {
  // TODO: is this the best query?
  const response = await requestWithOrgID('/querier.v1.QuerierService/Series', {
    method: 'POST',
    body: JSON.stringify({
      matchers: [],
    }),
    headers: {
      'content-type': 'application/json',
    },
  });

  if (response.isOk) {
    return parseResponse(response, AppsSchema);
  }

  return Result.err<App[], RequestError>(response.error);
}
