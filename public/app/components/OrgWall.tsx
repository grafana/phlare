import React, { useEffect, useState } from 'react';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import TextField from '@webapp/ui/Form/TextField';
import {
  Dialog,
  DialogBody,
  DialogFooter,
  DialogHeader,
} from '@webapp/ui/Dialog';
import Button from '@webapp/ui/Button';
import {
  checkTenancyIsRequired,
  selectTenancy,
  actions,
  selectOrgID,
} from '../redux/reducers/org';

/*
 * OrgWall checks whether the user is running in a multitenant environment
 * and if so, asks for an org to be set
 * which is then stored in localStorage
 */
export function OrgWall({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch();
  const tenancy = useAppSelector(selectTenancy);
  const currentOrg = useAppSelector(selectOrgID);

  useEffect(() => {
    void dispatch(checkTenancyIsRequired());
  }, [dispatch]);

  switch (tenancy) {
    case 'unknown':
    case 'loading': {
      return <></>;
    }
    case 'wants_to_change': {
      return (
        <>
          <SelectOrgDialog
            currentOrg={currentOrg}
            onSaved={(orgID) => {
              console.log('setting orgId', orgID);
              void dispatch(actions.setOrgID(orgID));
            }}
          />
          {children}
        </>
      );
    }
    case 'needs_org_id': {
      return (
        <SelectOrgDialog
          currentOrg={currentOrg}
          onSaved={(orgID) => {
            console.log('setting orgId', orgID);
            void dispatch(actions.setOrgID(orgID));
          }}
        />
      );
    }
    case 'multi_tenant':
    case 'single_tenant': {
      return <>{children}</>;
    }
  }
}

function SelectOrgDialog({
  currentOrg,
  onSaved,
}: {
  currentOrg?: string;
  onSaved: (orgID: string) => void;
}) {
  const [isDialogOpen] = useState(true);
  const handleFormSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    console.log('submitting form', e.target);
    const orgID = e.target.orgID.value;
    onSaved(orgID);
  };

  return (
    <>
      <Dialog open={isDialogOpen} aria-labelledby="dialog-header">
        <>
          <DialogHeader>
            <h3 id="dialog-header">Enter an Organization ID</h3>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              void handleFormSubmit(e);
            }}
          >
            <DialogBody>
              <>
                <p>
                  Your instance has been detected as a multitenant one. Please
                  enter an Organization ID (You can change it at any time via
                  the sidebar).
                </p>

                <TextField
                  defaultValue={currentOrg}
                  label="Organization ID"
                  required
                  id="orgID"
                  name="displayName"
                  type="text"
                  autoFocus
                />
              </>
            </DialogBody>
            <DialogFooter>
              <Button type="submit" kind="secondary">
                Submit
              </Button>
            </DialogFooter>
          </form>
        </>
      </Dialog>
    </>
  );
}
