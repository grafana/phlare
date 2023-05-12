import { fetchApps } from '@webapp/services/apps';
import { RequestNotOkError } from '@webapp/services/base';
import store from '@phlare/redux/store';

export const LOCAL_STORAGE_PREFIX = 'pyroscope:tenant';

export async function isMultiTenancyEnabled() {
  const res = await fetchApps();
  return isOrgRequiredError(res);
}

function isOrgRequiredError(res: Awaited<ReturnType<typeof fetchApps>>) {
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
