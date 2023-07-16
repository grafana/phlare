import { parseResponse, requestWithOrgID } from '@webapp/services/base';
import { z } from 'zod';

const labelNamesSchema = z.preprocess(
  (a: any) => {
    if ('names' in a) {
      return a;
    }
    return { names: [] };
  },
  z.object({
    names: z.array(z.string()),
  })
);

// todo we  should filter using matchers = [service_name="foo", tag="bar", etc]
export async function fetchTags(query: string, _from: number, _until: number) {
  const profileTypeID = query.replace(/\{.*/g, '');
  const response = await requestWithOrgID(
    '/querier.v1.QuerierService/LabelNames',
    {
      method: 'POST',
      body: JSON.stringify({
        matchers: [`{__profile_type__=\"${profileTypeID}\"}`],
      }),
      headers: {
        'content-type': 'application/json',
      },
    }
  );
  const isMetaTag = (tag: string) => tag.startsWith('__') && tag.endsWith('__');

  return parseResponse<string[]>(
    response,
    labelNamesSchema.transform((res)  =>  {
      return Array.from(new Set(res.names.filter((a) => !isMetaTag(a))));
    })
  );
}

// todo we  should filter using matchers = [service_name="foo", tag="bar", etc]
export async function fetchLabelValues(
  label: string,
  query: string,
  _from: number,
  _until: number
) {
  const profileTypeID = query.replace(/\{.*/g, '');
  const response = await requestWithOrgID(
    '/querier.v1.QuerierService/LabelValues',
    {
      method: 'POST',
      body: JSON.stringify({
        matchers: [`{__profile_type__=\"${profileTypeID}\"}`],
        name: label,
      }),
      headers: {
        'content-type': 'application/json',
      },
    }
  );

  return parseResponse<string[]>(
    response,
    labelNamesSchema.transform((res) => {
      return Array.from(new Set(res.names));
    })
  );
}
