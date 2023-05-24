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
