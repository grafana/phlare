import { parse, brandQuery, Query } from '@webapp/models/query';
import { z } from 'zod';

// TODO(eh-am): update to __service_name__ after https://github.com/grafana/phlare/pull/710 is merged
export const AppNameLabel = 'pyroscope_app' as const;
export type AppNameLabelType = typeof AppNameLabel;

export const AppSchema = z.object({
  __profile_type__: z.string(),
  [AppNameLabel]: z.string(),

  // TODO: add more fields (like spyName)
  // TODO: validate using UnitsSchema
  //  __type__: z.string(),
  //  query: z.string(),
  //  TODO: this field is currently only used as a sortKey in redux
  name: z.string().optional().default(''),
});

export type App = z.infer<typeof AppSchema>;

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

export function appToQuery(
  app: Pick<App, '__profile_type__' | AppNameLabelType>
): Query {
  return brandQuery(
    `${app.__profile_type__}{${AppNameLabel}="${app[AppNameLabel]}"}`
  );
}
