const KEY_NAME = 'PYROSCOPE_ORG_ID';

export function setOrgID(orgID: string) {
  localStorage.setItem(KEY_NAME, orgID);
}

export function getOrgID() {
  return localStorage.getItem(KEY_NAME);
}
