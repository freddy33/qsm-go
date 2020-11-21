import React from 'react';
import { Router } from '@reach/router';

import PathContextList from './PathContextList';
import RenderPage from './RenderPage';
import EnvironmentPage from './EnvironmentPage';

const NotFound = () => <h1>Invalid route</h1>;

const App = () => {
  return (
    <Router>
      <PathContextList path="/" />
      <PathContextList path="path-contexts" />
      <RenderPage path="render" />
      <RenderPage path="render/:pathContextId" />
      <EnvironmentPage path="environments" />

      <NotFound default />
    </Router>
  );
};

export default App;
