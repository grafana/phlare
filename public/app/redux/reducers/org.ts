import { createSlice, type PayloadAction } from '@reduxjs/toolkit';
import { createAsyncThunk } from '@webapp/redux/async-thunk';
import type { RootState } from '../store';
import {
  LOCAL_STORAGE_PREFIX,
  isMultiTenancyEnabled,
  tenantIDFromStorage,
} from '@phlare/services/tenant';
import storage from 'redux-persist/lib/storage';
import { PersistConfig } from 'redux-persist/lib/types';

export const persistConfig: PersistConfig<OrgState> = {
  key: LOCAL_STORAGE_PREFIX,
  version: 0,
  storage,
  whitelist: ['orgID'],
};

interface OrgState {
  tenancy:
    | 'unknown'
    | 'loading'
    | 'needs_org_id'
    | 'wants_to_change'
    | 'single_tenant'
    | 'multi_tenant';
  orgID?: string;
}

const initialState: OrgState = {
  tenancy: 'unknown',
  orgID: undefined,
};

export const checkTenancyIsRequired = createAsyncThunk<
  { tenancy: OrgState['tenancy']; orgID?: string },
  void,
  { state: { org: OrgState } }
>(
  'checkTenancyIsRequired',
  async () => {
    const orgID = tenantIDFromStorage();

    // Try to hit the server and see the response
    const multitenancy = await isMultiTenancyEnabled();

    if (multitenancy && !orgID) {
      return Promise.resolve({ tenancy: 'needs_org_id', orgID });
    }

    if (multitenancy && orgID) {
      return Promise.resolve({ tenancy: 'multi_tenant', orgID });
    }

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
      state.tenancy = 'unknown';
      state.orgID = undefined;
    },
    setOrgID(state, action: PayloadAction<string>) {
      state.tenancy = 'multi_tenant';
      state.orgID = action.payload;
    },
    setWantsToChange(state) {
      state.tenancy = 'wants_to_change';
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
  state.org.tenancy === 'multi_tenant' ||
  state.org.tenancy === 'wants_to_change';

export const selectOrgID = (state: RootState) => state.org.orgID;

export default orgSlice.reducer;
