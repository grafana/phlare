import { parse, brandQuery, Query } from '@webapp/models/query';
import { z } from 'zod';

export const AppSchema = z.object({
  __profile_type__: z.string(),
  pyroscope_app: z.string(),

  // TODO: add more fields (like spyName)
  // TODO: validate using UnitsSchema
  //  __type__: z.string(),
  //  query: z.string(),
  //  TODO: this field is currently only used for sorting
  name: z.string().optional(),
});

export type App = z.infer<typeof AppSchema>;

// TODO(eh-am): update to __service_name__
export const AppNameLabel = 'pyroscope_app';

// Given a query returns an App
export function appFromQuery(query: Query): App | undefined {
  const parsed = parse(query);

  if (!parsed) {
    return undefined;
  }

  const app = {
    __profile_type__: parsed?.profileId,
    ...parsed?.tags,
  };

  const parsedApp = AppSchema.safeParse(app);
  if (!parsedApp.success) {
    return undefined;
  }

  return parsedApp.data;
}

export function appToQuery(app: App): Query {
  return brandQuery(
    `${app.__profile_type__}{${AppNameLabel}="${app.pyroscope_app}"}`
  );
}
