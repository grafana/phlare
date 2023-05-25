//export * from '@pyroscope/webapp/javascript/components/AppSelector';
import React, { useState, useEffect, useMemo } from 'react';
import ModalWithToggle from '@webapp/ui/Modals/ModalWithToggle';
import Input from '@webapp/ui/Input';
import { App } from '@webapp/models/app';
import { parse, brandQuery, Query } from '@webapp/models/query';
import cx from 'classnames';
//import SelectButton from '@webapp/components/AppSelector/SelectButton';
// TODO:
import SelectButton from '@pyroscope/webapp/javascript/components/AppSelector/SelectButton';
//import { Label, LabelString } from '@webapp/components/AppSelector/Label';
//import styles from '@pyroscope/webapp/AppSelector.module.scss';
// TODO:
import styles from '../../../../node_modules/pyroscope-oss/webapp/javascript/components/AppSelector/AppSelector.module.scss';
import styles2 from './AppSelector.module.css';

interface AppSelectorProps {
  /** Triggered when an app is selected */
  onSelected: (query: Query) => void;

  /** List of all applications */
  apps: App[];

  selectedQuery: string;
}

/**
 * Given a flat list of Apps
 * Get unique app names
 */
function getAppNames(apps: App[]) {
  const names = apps.map((a) => {
    return a.pyroscope_app;
  });

  return Array.from(new Set(names));
}

function findAppsWithName(apps: App[], appName: string) {
  return apps.filter((a) => {
    return a.pyroscope_app === appName;
  });
}

export function AppSelector({
  onSelected,
  apps,
  selectedQuery,
}: AppSelectorProps) {
  return (
    <div className={styles.container}>
      <SelectorModalWithToggler
        apps={apps}
        onSelected={(app) => onSelected(appToQuery(app))}
        selectedQuery={selectedQuery}
      />
    </div>
  );
}

function appToQuery(app: App) {
  return brandQuery(
    `${app.__profile_type__}{pyroscope_app="${app.pyroscope_app}"}`
  );
}

export const SelectorModalWithToggler = ({
  apps,
  onSelected,
}: {
  apps: App[];
  onSelected: (app: App) => void;
}) => {
  const appNames = getAppNames(apps);
  //  const [filter, setFilter] = useState('');
  const [isModalOpen, setModalOpenStatus] = useState(false);
  // TODO: name
  const [selectedAppName, setSelectedAppName] = useState<string>();

  // TODO: use memo
  const matchedApps = findAppsWithName(apps, selectedAppName || '');

  return (
    <ModalWithToggle
      isModalOpen={isModalOpen}
      setModalOpenStatus={setModalOpenStatus}
      modalClassName={cx(styles.appSelectorModal, styles2.appSelectorModal)}
      modalHeight={'auto'}
      noDataEl={
        !appNames?.length ? (
          <div data-testid="app-selector-no-data" className={styles.noData}>
            No Data
          </div>
        ) : null
      }
      toggleText={'TEST'}
      headerEl={
        <></> && (
          <>
            <div className={styles.headerTitle}>{'TEST'}</div>
            <Input
              name="application search"
              type="text"
              placeholder="Type an app"
              value={''}
              onChange={''}
              className={styles.search}
              testId="application-search"
            />
          </>
        )
      }
      leftSideEl={appNames.map((name) => (
        <SelectButton
          name={name}
          onClick={() => {
            setSelectedAppName(name);
          }}
          fullList={appNames}
          isSelected={selectedAppName === name}
          key={name}
        />
      ))}
      rightSideEl={matchedApps.map((app) => (
        <SelectButton
          name={app.__profile_type__}
          onClick={() => onSelected(app)}
          fullList={appNames}
          isSelected={false}
          key={app.__profile_type__}
        />
      ))}
    />
  );
};
//
//
export default AppSelector;
