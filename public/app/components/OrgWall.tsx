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

export function OrgWall({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch();
  const tenancy = useAppSelector(selectTenancy);
  const currentOrg = useAppSelector(selectOrgID);

  useEffect(() => {
    void dispatch(checkTenancyIsRequired());
  }, [dispatch]);

  // Don't rerender all the children when this component changes
  // For example, when user wants to change the tenant ID
  const memoedChildren = React.useMemo(() => children, []);

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
              void dispatch(actions.setOrgID(orgID));
            }}
          />
          {memoedChildren}
        </>
      );
    }
    case 'needs_org_id': {
      return (
        <SelectOrgDialog
          currentOrg={currentOrg}
          onSaved={(orgID) => {
            void dispatch(actions.setOrgID(orgID));
          }}
        />
      );
    }
    case 'multi_tenant':
    case 'single_tenant': {
      return <>{memoedChildren}</>;
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

    const target = e.target as typeof e.target & {
      tenantID: { value: string };
    };

    onSaved(target.tenantID.value);
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
                  enter a Tenant ID (You can change it at any time via the
                  sidebar).
                </p>
                <p>
                  Notice that if you migrated from the not multi-tenant version
                  the data can be found under tenant ID "anonymous".
                </p>

                <TextField
                  defaultValue={currentOrg}
                  label="Tenant ID"
                  required
                  id="tenantID"
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
