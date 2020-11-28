import React from 'react';
import { Link } from '@reach/router';
import styles from './index.module.scss';

const NavLink = (props) => (
  <Link
    {...props}
    getProps={({ isCurrent }) => {
      return {
        className: isCurrent ? styles.active : null,
      };
    }}
  />
);

const PageHeader = ({ currentEnv }) => {
  return (
    <div className={styles.pageHeader}>
      <nav className={styles.breadcrumb}>
        <NavLink to="/">Environments</NavLink>
        {' > '}
        <NavLink to="/path-contexts">Path Contexts</NavLink>
        {' > '}
        <NavLink to="/render">Render</NavLink>
      </nav>
      <div className={styles.currentEnv}>(Current Env: {currentEnv})</div>
    </div>
  );
};

export default PageHeader;
