import { Result } from '@webapp/util/fp';
import {
  Profile,
  Groups,
  FlamebearerProfileSchema,
  GroupsSchema,
} from '@pyroscope/models/src';
import { z } from 'zod';
import type { ZodError } from 'zod';
import { buildRenderURL } from '@webapp/util/updateRequests';
import { Timeline, TimelineSchema } from '@webapp/models/timeline';
import { Annotation, AnnotationSchema } from '@webapp/models/annotation';
import type { RequestError } from '@webapp/services/base';
import { parseResponse, requestWithOrgID } from '@webapp/services/base';

export interface RenderOutput {
  profile: Profile;
  timeline: Timeline;
  groups?: Groups;
  annotations: Annotation[];
}

// Default to empty array if not present
const defaultAnnotationsSchema = z.preprocess((a) => {
  if (!a) {
    return [];
  }
  return a;
}, z.array(AnnotationSchema));

interface RenderSingleProps {
  from: string;
  until: string;
  query: string;
  refreshToken?: string;
  maxNodes: string | number;
}
export async function renderSingle(
  props: RenderSingleProps,
  controller?: {
    signal?: AbortSignal;
  }
): Promise<Result<RenderOutput, RequestError | ZodError>> {
  const url = buildRenderURL(props);
  // TODO
  const response = await requestWithOrgID(`/pyroscope/${url}&format=json`, {
    signal: controller?.signal,
  });

  if (response.isErr) {
    return Result.err<RenderOutput, RequestError>(response.error);
  }

  const parsed = FlamebearerProfileSchema.merge(
    z.object({
      timeline: TimelineSchema,
      annotations: defaultAnnotationsSchema,
    })
  )
    .merge(z.object({ telemetry: z.object({}).passthrough().optional() }))
    .safeParse(response.value);

  if (parsed.success) {
    // TODO: strip timeline
    const profile = parsed.data;
    const { timeline, annotations } = parsed.data;

    return Result.ok({
      profile,
      timeline,
      annotations,
    });
  }

  return Result.err(parsed.error);
}

export type RenderDiffResponse = z.infer<typeof FlamebearerProfileSchema>;

interface RenderDiffProps {
  leftFrom: string;
  leftUntil: string;
  leftQuery: string;
  refreshToken?: string;
  maxNodes: string;
  rightQuery: string;
  rightFrom: string;
  rightUntil: string;
}

export async function renderDiff(
  props: RenderDiffProps,
  controller?: {
    signal?: AbortSignal;
  }
) {
  const params = new URLSearchParams({
    leftQuery: props.leftQuery,
    leftFrom: props.leftFrom,
    leftUntil: props.leftUntil,
    rightQuery: props.rightQuery,
    rightFrom: props.rightFrom,
    rightUntil: props.rightUntil,
  });

  const response = await requestWithOrgID(`/pyroscope/render-diff?${params}`, {
    signal: controller?.signal,
  });

  return parseResponse<z.infer<typeof FlamebearerProfileSchema>>(
    response,
    FlamebearerProfileSchema
  );
}

const RenderExploreSchema = FlamebearerProfileSchema.extend({
  groups: z.preprocess((groups) => {
    const groupNames = Object.keys(groups as Groups);
    return groupNames.length
      ? groupNames
          .filter((g) => !!g.trim())
          .reduce(
            (acc, current) => ({
              ...acc,
              [current]: (groups as Groups)[current],
            }),
            {}
          )
      : groups;
  }, GroupsSchema),
}).transform((values) => {
  return {
    profile: values,
    groups: values.groups,
  };
});

interface RenderExploreProps extends Omit<RenderSingleProps, 'maxNodes'> {
  groupBy: string;
  grouByTagValue: string;
}

export type RenderExploreOutput = z.infer<typeof RenderExploreSchema>;

export async function renderExplore(
  props: RenderExploreProps,
  controller?: {
    signal?: AbortSignal;
  }
): Promise<Result<RenderExploreOutput, RequestError | ZodError>> {
  const url = buildRenderURL(props);
  const response = await requestWithOrgID(`/pyroscope/${url}&format=json`, {
    signal: controller?.signal,
  });
  return parseResponse<RenderExploreOutput>(response, RenderExploreSchema);
}
