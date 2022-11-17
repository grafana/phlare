import { DataQueryRequest, DataQueryResponse, DataSourceInstanceSettings } from '@grafana/data';
import { DataSourceWithBackend } from '@grafana/runtime';
import { PhlareDataSourceOptions, Query, ProfileTypeMessage, SeriesMessage } from './types';
import { Observable, of } from 'rxjs';
import { normalizeQuery } from './QueryEditor/QueryEditor';

export class PhlareDataSource extends DataSourceWithBackend<Query, PhlareDataSourceOptions> {
  constructor(instanceSettings: DataSourceInstanceSettings<PhlareDataSourceOptions>) {
    super(instanceSettings);
  }

  query(request: DataQueryRequest<Query>): Observable<DataQueryResponse> {
    const validTargets = request.targets
      .filter((t) => t.profileTypeId)
      .map((t) => {
        // Empty string errors out but honestly seems like we can just normalize it this way
        if (t.labelSelector === '') {
          return {
            ...t,
            labelSelector: '{}',
          };
        }

        return normalizeQuery(t, request.app);
      });

    if (!validTargets.length) {
      return of({ data: [] });
    }
    return super.query({
      ...request,
      targets: validTargets,
    });
  }

  async getProfileTypes(): Promise<ProfileTypeMessage[]> {
    return await super.getResource('profileTypes');
  }

  async getSeries(): Promise<SeriesMessage> {
    // For now, we send empty matcher to get all the series
    return await super.getResource('series', { matchers: ['{}'] });
  }

  async getLabelNames(): Promise<string[]> {
    return await super.getResource('labelNames');
  }
}
