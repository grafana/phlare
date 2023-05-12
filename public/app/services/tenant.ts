import { RequestNotOkError } from '@webapp/services/base';
import store from '@phlare/redux/store';
import { request } from '@webapp/services/base';

export const LOCAL_STORAGE_PREFIX = 'pyroscope:tenant';

export async function isMultiTenancyEnabled() {
  // Do a request not passing any headers
  const res = await request('/pyroscope/label-values?label=__name__', {
    headers: {
      'X-Scope-OrgID': '',
    },
  });

  if (res.isOk) {
    return false;
  }

  return isOrgRequiredError(res);
}

function isOrgRequiredError(res: Awaited<ReturnType<typeof request>>) {
  // TODO: is 'no org id' a stable message?
  return (
    res.isErr &&
    res.error instanceof RequestNotOkError &&
    res.error.code == 401 &&
    res.error.description === 'no org id\n'
  );
}

// The source of truth is actually from redux-persist
// However we may need to access directly from local storage
// Eg when doing a request
export function tenantIDFromStorage(): string {
  return store.getState().org.orgID || '';
}
