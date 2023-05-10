import { faCog } from '@fortawesome/free-solid-svg-icons/faCog';
import { faUser } from '@fortawesome/free-solid-svg-icons/faUser';
import { MenuButton, MenuProps, MenuHeader } from '@szhsin/react-menu';
import Dropdown, { MenuItem as DropdownMenuItem } from '@webapp/ui/Dropdown';
import flattenChildren from 'react-flatten-children';
import Icon from '@webapp/ui/Icon';
import { MenuItem } from '@webapp/ui/Sidebar';
import {
  selectIsMultiTenant,
  selectOrgID,
  actions,
} from '../redux/reducers/org';
import { useAppSelector, useAppDispatch } from '../redux/hooks';
import styles from './SidebarOrg.module.css';
import Button from '@webapp/ui/Button';

export interface DropdownProps {
  children: JSX.Element[] | JSX.Element;
  offsetX: MenuProps['offsetX'];
  offsetY: MenuProps['offsetY'];
  direction: MenuProps['direction'];
  label: string;
  className: string;
  menuButton: JSX.Element;
}

function FlatDropdown({
  children,
  offsetX,
  offsetY,
  direction,
  label,
  className,
  menuButton,
}: DropdownProps) {
  return (
    <Dropdown
      offsetX={offsetX}
      offsetY={offsetY}
      direction={direction}
      label={label}
      className={className}
      menuButton={menuButton}
    >
      {flattenChildren(children) as unknown as JSX.Element}
    </Dropdown>
  );
}

export default function AccountButton() {
  const isMultiTenant = useAppSelector(selectIsMultiTenant);
  const orgID = useAppSelector(selectOrgID);
  const dispatch = useAppDispatch();

  if (!isMultiTenant) {
    return <></>;
  }

  // TODO: show the modal
  const onChangeOrganizationClick = () => {
    dispatch(actions.deleteTenancy());

    window.location.reload();
  };

  return (
    <>
      <FlatDropdown
        offsetX={10}
        offsetY={5}
        direction="top"
        label=""
        className={styles.dropdown}
        menuButton={
          <MenuButton className={styles.accountDropdown}>
            <MenuItem icon={<Icon icon={faUser} />}>Tenancy</MenuItem>
          </MenuButton>
        }
      >
        <MenuHeader>Current OrgID</MenuHeader>
        <DropdownMenuItem className={styles.menuItemDisabled}>
          <div className={styles.menuItemWithButton}>
            <span className={styles.menuItemWithButtonTitle}>
              Org ID: {orgID}
            </span>
            <Button className={styles.menuItemWithButtonButton}>
              <div onClick={() => onChangeOrganizationClick()}>
                <Icon icon={faCog} />
              </div>
            </Button>
          </div>
        </DropdownMenuItem>
      </FlatDropdown>
    </>
  );
}
