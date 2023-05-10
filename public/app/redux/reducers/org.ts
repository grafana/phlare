import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import { createAsyncThunk } from '@webapp/redux/async-thunk';
import { fetchApps } from '@webapp/services/apps';
import type { RootState } from '../store';
import { RequestNotOkError } from '@webapp/services/base';
import { deleteOrgID, getOrgID, setOrgID } from '../../services/orgID';

interface OrgState {
  tenancy:
    | 'unknown'
    | 'loading'
    | 'needs_org_id'
    | 'single_tenant'
    | 'multi_tenant';
  orgID?: string;
}

const initialState: OrgState = {
  tenancy: 'unknown',
  orgID: undefined,
};

function isOrgRequiredError(res: Awaited<ReturnType<typeof fetchApps>>) {
  // TODO: is 'no org id' a stable message?
  return (
    res.isErr &&
    res.error instanceof RequestNotOkError &&
    res.error.code == 401 &&
    res.error.description === 'no org id\n'
  );
}

export const checkTenancyIsRequired = createAsyncThunk<
  { tenancy: OrgState['tenancy']; orgID?: string },
  void,
  { state: { org: OrgState } }
>(
  'checkTenancyIsRequired',
  async () => {
    const orgID = getOrgID() || undefined;
    // There's an orgID previously set
    if (orgID) {
      return Promise.resolve({
        tenancy: 'multi_tenant',
        orgID,
      });
    }

    // Try to hit the server and see the response
    const res = await fetchApps();
    if (isOrgRequiredError(res)) {
      return Promise.resolve({ tenancy: 'needs_org_id', orgID });
    }

    // No error, or it's a different kind of error
    // Let's assume it's a single tenant
    return Promise.resolve({ tenancy: 'single_tenant', orgID });
  },
  {
    // This check is only valid if we don't know what's the tenancy status yet
    condition: (query, thunkAPI) => {
      thunkAPI.getState;
      const state = thunkAPI.getState().org;

      return state.tenancy === 'unknown';
    },
  }
);

const orgSlice = createSlice({
  name: 'org',
  initialState,
  reducers: {
    deleteTenancy(state) {
      deleteOrgID();
      state.tenancy = 'unknown';
    },
    setOrgID(state, action: PayloadAction<string>) {
      setOrgID(action.payload);
      state.tenancy = 'multi_tenant';
    },
    setTenancy(state, action: PayloadAction<OrgState['tenancy']>) {
      state.tenancy = action.payload;
    },
  },
  extraReducers: (builder) => {
    // This thunk will never reject
    builder.addCase(checkTenancyIsRequired.fulfilled, (state, action) => {
      state.tenancy = action.payload.tenancy;
      state.orgID = action.payload.orgID;
    });
    builder.addCase(checkTenancyIsRequired.pending, (state) => {
      state.tenancy = 'loading';
    });
  },
});

export const { actions } = orgSlice;

export const selectTenancy = (state: RootState) => state.org.tenancy;

export const selectIsMultiTenant = (state: RootState) =>
  state.org.tenancy === 'multi_tenant';

export const selectOrgID = (state: RootState) => state.org.orgID;

export default orgSlice.reducer;
