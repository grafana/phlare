import { Link } from "react-router-dom";
import {
  Collapse,
  Navbar,
  NavbarBrand,
  NavbarText,
  NavbarToggler,
  Nav,
  NavItem,
  NavLink,
  UncontrolledDropdown,
  DropdownToggle,
  DropdownMenu,
  DropdownItem,
} from 'reactstrap';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faFire } from '@fortawesome/free-solid-svg-icons';


export default function Navigation() {
  return(<Navbar
    color="light"
    expand="md"
    light
  >
    <NavbarBrand href="/">
      Grafana Fire <FontAwesomeIcon icon={faFire} />
    </NavbarBrand>
    <NavbarToggler onClick={function noRefCheck(){}} />
    <Collapse navbar>
      <Nav
        className="me-auto"
        navbar
      >
        <NavItem>
          <NavLink href="/query">
            Query
          </NavLink>
        </NavItem>
        <NavItem>
          <NavLink href="https://github.com/grafana/fire">
            GitHub
          </NavLink>
        </NavItem>
        <UncontrolledDropdown
          inNavbar
          nav
        >
          <DropdownToggle
            caret
            nav
          >
            Help
          </DropdownToggle>
          <DropdownMenu right>
            <DropdownItem tag={Link} to="/documentation">
              Documentation
            </DropdownItem>
            <DropdownItem tag={Link} to="/api">
              API
            </DropdownItem>
          </DropdownMenu>
        </UncontrolledDropdown>
      </Nav>
      <NavbarText>
        Continous Profiling
      </NavbarText>
    </Collapse>
  </Navbar>
)
}
