import { brandQuery, Query } from '@webapp/models/query';
import { z } from 'zod';

export const AppSchema = z.object({
  __profile_type__: z.string(),
  pyroscope_app: z.string(),

  // TODO: add more fields (like spyName)
  // TODO: validate using UnitsSchema
  __type__: z.string(),
  query: z.string(),
  name: z.string(),
});

export type App = z.infer<typeof AppSchema>;

// TODO(eh-am): update to __service_name__
export const AppNameLabel = 'pyroscope_app';

export function appFromQuery(apps: App[], query: Query) {
  // TODO: from the query, parse
  // - a profileID
  // - a name
}

export function appToQuery(app: App): Query {
  return brandQuery(
    `${app.__profile_type__}{${AppNameLabel}="${app.pyroscope_app}"}`
  );
}
