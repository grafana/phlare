import { Result } from '@webapp/util/fp';
import {
  type RequestError,
  request as ogRequest,
} from '../../../../node_modules/pyroscope-oss/webapp/javascript/services/base';
import { getOrgID } from '../../services/orgID';

export * from '../../../../node_modules/pyroscope-oss/webapp/javascript/services/base';

/**
 * request wraps around the original request
 * while sending the OrgID if available
 */
export async function request(
  request: RequestInfo,
  config?: RequestInit
): Promise<Result<unknown, RequestError>> {
  const headers = {
    ...config?.headers,
    // TODO: fetching from localstorage every time may be slow
    'X-Scope-OrgID': getOrgID() || '',
  };

  return ogRequest(request, {
    ...config,
    headers,
  });
}
