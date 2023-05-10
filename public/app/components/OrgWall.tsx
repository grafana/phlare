import React, { useEffect, useState } from 'react';
import { fetchApps } from '@webapp/services/apps';
import { useAppDispatch, useAppSelector } from '../redux/hooks';
import { RequestNotOkError } from '@webapp/services/base';
import TextField from '@webapp/ui/Form/TextField';
import {
  Dialog,
  DialogBody,
  DialogFooter,
  DialogHeader,
} from '@webapp/ui/Dialog';
import Button from '@webapp/ui/Button';
import { setOrgID } from '../services/orgID';

type States = 'INITIAL' | 'NEEDS_ORG' | 'ALL_GOOD';

/*
 * OrgWall checks whether the user is running in a multitenant environment
 * and if so, asks for an org to be set
 * which is then stored in localStorage
 */
export function OrgWall({ children }: { children: React.ReactNode }) {
  const dispatch = useAppDispatch();
  const [currentState, setCurrentState] = useState<States>('INITIAL');

  useEffect(() => {
    async function run() {
      // We do the request directly via the service
      // Without dispatching an action
      // Since actions with errors are globally handled via the Notifications system
      const res = await fetchApps();

      if (isOrgRequiredError(res)) {
        // We are sure error is due to lack of orgID
        // So let's show the modal
        setCurrentState('NEEDS_ORG');
        return;
      }

      // Any other kind of error
      // Or everything went fine
      // Let's let the rest of the application handle it
      setCurrentState('ALL_GOOD');
    }

    run();
  }, [dispatch]);

  switch (currentState) {
    case 'INITIAL': {
      return <></>;
    }
    case 'NEEDS_ORG': {
      return (
        <SelectOrgDialog
          onSaved={() => {
            setCurrentState('ALL_GOOD');
          }}
        />
      );
    }
    case 'ALL_GOOD': {
      return <>{children}</>;
    }
  }
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

function SelectOrgDialog({ onSaved }: { onSaved: () => void }) {
  const [isDialogOpen] = useState(true);
  const handleFormSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    const orgID = e.target.orgID.value;
    setOrgID(orgID);
    onSaved();
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
