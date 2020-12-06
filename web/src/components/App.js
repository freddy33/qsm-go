import React, { useState } from 'react';
import { Router } from '@reach/router';

import PathContextList from './PathContextList';
import RenderPage from './RenderPage';
import EnvironmentPage from './EnvironmentPage';
import PageHeader from './shared/PageHeader';
import styles from './App.module.scss';
import localStorage from '../libs/util/localStorage';
import { LOCAL_STORAGE_KEY } from '../libs/constant';
import SpacePage from './SpacePage';
import EventPage from './EventPage';

const NotFound = () => <h1>Invalid route</h1>;

const App = () => {
  const [currentEnv, setCurrentEnv] = useState(localStorage.getCurrentEnv());

  const changeEnv = (envId) => {
    localStorage.setItem(LOCAL_STORAGE_KEY.SELECTED_ENVIRONMENT, envId);
    setCurrentEnv(envId);
  };

  return (
    <div className={styles.app}>
      <PageHeader currentEnv={currentEnv} />
      <Router className={styles.content}>
        <EnvironmentPage path="/" changeEnv={changeEnv} />
        <PathContextList path="path-contexts" />
        <RenderPage path="render" />
        <RenderPage path="render/:pathContextId" />
        <EnvironmentPage path="environments" />
        <SpacePage path="spaces" />
        <EventPage path="events/:spaceId" />

        <NotFound default />
      </Router>
    </div>
  );
};

export default App;
