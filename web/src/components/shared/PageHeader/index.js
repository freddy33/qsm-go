import React from 'react';
import { Link } from '@reach/router';
import styles from './index.module.scss';

const NavLink = (props) => (
  <Link
    {...props}
    getProps={({ isPartiallyCurrent, isCurrent }) => {
      const active =
        (props.partialMatch && isPartiallyCurrent) ||
        (!props.partialMatch && isCurrent);
      return {
        className: active ? styles.active : null,
      };
    }}
  />
);

const PageHeader = ({ currentEnv }) => {
  return (
    <div className={styles.pageHeader}>
      <nav className={styles.breadcrumb}>
        <NavLink to="/">Environments</NavLink>
        {' | '}
        <NavLink to="/spaces">Spaces</NavLink>
        {' | '}
        <NavLink to="/path-contexts">Path Contexts</NavLink>
        {' | '}
        <NavLink to="/render" partialMatch>
          Render
        </NavLink>
      </nav>
      <div className={styles.currentEnv}>(Current Env: {currentEnv})</div>
    </div>
  );
};

export default PageHeader;
