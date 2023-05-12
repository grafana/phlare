import { RequestNotOkError } from '@webapp/services/base';
import store from '@phlare/redux/store';
import { request } from '@webapp/services/base';

export async function isMultiTenancyEnabled() {
  // Do a request not passing any headers
  const res = await request('/pyroscope/label-values?label=__name__', {
    headers: {
      'X-Scope-OrgID': '',
    },
  });

  // If everything went okay even without passing an OrgID, we can assume it's a non multitenant instance
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

export function tenantIDFromStorage(): string {
  return store.getState().tenant.tenantID || '';
}
